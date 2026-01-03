package service

import (
	"context"
	"fmt"

	"github.com/ChotongW/grit_demo_wallet/internal/accounts/repository"
	pbSub "github.com/ChotongW/grit_demo_wallet/pb/subledger"

	accountErrors "github.com/ChotongW/grit_demo_wallet/internal/accounts/errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

const (
	ReferralFundingPoolAccount     = "1001"
	InstitutionMainAccount         = "1002"
	InstitutionDisbursementAccount = "1003"
	PSPAccount                     = "1004"
	ReferralRewardAmount           = "10.00"
)

type Service struct {
	repo            *repository.Repository
	subledgerClient pbSub.SubledgerServiceClient
	logger          *logrus.Entry
}

func NewService(repo *repository.Repository, subledgerClient pbSub.SubledgerServiceClient, logger *logrus.Logger) *Service {
	return &Service{
		repo:            repo,
		subledgerClient: subledgerClient,
		logger: logger.WithFields(logrus.Fields{
			"package": "accounts/service",
		}),
	}
}

func (s *Service) CreateAccount(ctx context.Context, email string, initialBalance decimal.Decimal, referrerAccountID string) (*repository.Account, error) {
	if referrerAccountID != "" {
		exists, err := s.repo.AccountExists(ctx, referrerAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate referrer: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("%w: %s", accountErrors.ErrInvalidReferrer, referrerAccountID)
		}
	}

	account, err := s.repo.CreateAccount(ctx, email, referrerAccountID)
	if err != nil {
		return nil, err
	}

	if initialBalance.GreaterThan(decimal.Zero) {
		refID := fmt.Sprintf("initial-deposit-%s", account.AccountID)
		desc := fmt.Sprintf("Initial deposit for account %s", account.AccountID)

		_, err = s.subledgerClient.CreateTransaction(ctx, &pbSub.CreateTransactionRequest{
			ReferenceId: refID,
			Description: desc,
			Entries: []*pbSub.Entry{
				{
					AccountId: PSPAccount,
					Amount:    initialBalance.String(),
					Direction: "DEBIT",
				},
				{
					AccountId: account.AccountID,
					Amount:    initialBalance.String(),
					Direction: "CREDIT",
				},
			},
		})

		if err != nil {
			s.logger.Errorf("Failed to create initial balance transaction: %v", err)
			return nil, fmt.Errorf("failed to create initial balance: %w", err)
		}

		s.logger.Infof("Created initial balance %s for account %s", initialBalance.String(), account.AccountID)
	}

	if referrerAccountID != "" {
		err = s.giveReferralReward(ctx, referrerAccountID, account.AccountID)
		if err != nil {
			s.logger.Errorf("Failed to give referral reward: %v", err)
		}
	}

	return account, nil
}

func (s *Service) giveReferralReward(ctx context.Context, referrerAccountID, newAccountID string) error {
	rewardAmount, _ := decimal.NewFromString(ReferralRewardAmount)
	refID := fmt.Sprintf("referral-reward-%s-%s", referrerAccountID, newAccountID)
	desc := fmt.Sprintf("Referral reward for referring account %s", newAccountID)

	_, err := s.subledgerClient.CreateTransaction(ctx, &pbSub.CreateTransactionRequest{
		ReferenceId: refID,
		Description: desc,
		Entries: []*pbSub.Entry{
			{
				AccountId: ReferralFundingPoolAccount,
				Amount:    rewardAmount.String(),
				Direction: "DEBIT",
			},
			{
				AccountId: referrerAccountID,
				Amount:    rewardAmount.String(),
				Direction: "CREDIT",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create referral reward transaction: %w", err)
	}

	s.logger.Infof("Gave referral reward %s to account %s", rewardAmount.String(), referrerAccountID)
	return nil
}

func (s *Service) GetAccount(ctx context.Context, accountID string) (*repository.Account, decimal.Decimal, error) {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		s.logger.Errorf("err: %+v", err)
		return nil, decimal.Zero, err
	}

	balance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return nil, decimal.Zero, err
	}

	return account, balance, nil
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	exists, err := s.repo.AccountExists(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}
	if !exists {
		return decimal.Zero, fmt.Errorf("%w: %s", accountErrors.ErrAccountNotFound, accountID)
	}

	return s.repo.GetBalance(ctx, accountID)
}

func (s *Service) Deposit(ctx context.Context, accountID string, amount decimal.Decimal, description string) (string, decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", decimal.Zero, accountErrors.ErrDepositAmountMustBePositive
	}
	exists, err := s.repo.AccountExists(ctx, accountID)
	if err != nil {
		return "", decimal.Zero, err
	}
	if !exists {
		return "", decimal.Zero, fmt.Errorf("%w: %s", accountErrors.ErrAccountNotFound, accountID)
	}

	refID := fmt.Sprintf("deposit-%s-%s", accountID, uuid.New().String())
	if description == "" {
		description = fmt.Sprintf("Deposit to account %s", accountID)
	}

	resp, err := s.subledgerClient.CreateTransaction(ctx, &pbSub.CreateTransactionRequest{
		ReferenceId: refID,
		Description: description,
		Entries: []*pbSub.Entry{
			{
				AccountId: PSPAccount,
				Amount:    amount.String(),
				Direction: "DEBIT",
			},
			{
				AccountId: accountID,
				Amount:    amount.String(),
				Direction: "CREDIT",
			},
		},
	})

	if err != nil {
		return "", decimal.Zero, fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	newBalance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return resp.TransactionId, decimal.Zero, err
	}

	s.logger.Infof("Deposited %s to account %s, new balance: %s", amount.String(), accountID, newBalance.String())
	return resp.TransactionId, newBalance, nil
}

func (s *Service) Withdraw(ctx context.Context, accountID string, amount decimal.Decimal, description string) (string, decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", decimal.Zero, accountErrors.ErrWithdrawAmountMustBePositive
	}
	currentBalance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return "", decimal.Zero, err
	}

	if currentBalance.LessThan(amount) {
		return "", decimal.Zero, fmt.Errorf("%w: have %s, need %s", accountErrors.ErrInsufficientBalance, currentBalance.String(), amount.String())
	}

	refID := fmt.Sprintf("withdraw-%s-%s", accountID, uuid.New().String())
	if description == "" {
		description = fmt.Sprintf("Withdrawal from account %s", accountID)
	}

	resp, err := s.subledgerClient.CreateTransaction(ctx, &pbSub.CreateTransactionRequest{
		ReferenceId: refID,
		Description: description,
		Entries: []*pbSub.Entry{
			{
				AccountId: accountID,
				Amount:    amount.String(),
				Direction: "DEBIT",
			},
			{
				AccountId: PSPAccount,
				Amount:    amount.String(),
				Direction: "CREDIT",
			},
		},
	})

	if err != nil {
		return "", decimal.Zero, fmt.Errorf("failed to create withdrawal transaction: %w", err)
	}

	newBalance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return resp.TransactionId, decimal.Zero, err
	}

	s.logger.Infof("Withdrew %s from account %s, new balance: %s", amount.String(), accountID, newBalance.String())
	return resp.TransactionId, newBalance, nil
}

func (s *Service) Transfer(ctx context.Context, fromAccountID, toAccountID string, amount decimal.Decimal, description string) (string, decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", decimal.Zero, accountErrors.ErrTransferAmountMustBePositive
	}

	if fromAccountID == toAccountID {
		return "", decimal.Zero, accountErrors.ErrTransferToSameAccount
	}

	fromExists, err := s.repo.AccountExists(ctx, fromAccountID)
	if err != nil {
		return "", decimal.Zero, err
	}
	if !fromExists {
		return "", decimal.Zero, fmt.Errorf("source %w: %s", accountErrors.ErrAccountNotFound, fromAccountID)
	}

	toExists, err := s.repo.AccountExists(ctx, toAccountID)
	if err != nil {
		return "", decimal.Zero, err
	}
	if !toExists {
		return "", decimal.Zero, fmt.Errorf("destination %w: %s", accountErrors.ErrAccountNotFound, toAccountID)
	}

	currentBalance, err := s.repo.GetBalance(ctx, fromAccountID)
	if err != nil {
		return "", decimal.Zero, err
	}

	if currentBalance.LessThan(amount) {
		return "", decimal.Zero, fmt.Errorf("%w: have %s, need %s", accountErrors.ErrInsufficientBalance, currentBalance.String(), amount.String())
	}

	refID := fmt.Sprintf("transfer-%s-%s-%s", fromAccountID, toAccountID, uuid.New().String())
	if description == "" {
		description = fmt.Sprintf("Transfer from %s to %s", fromAccountID, toAccountID)
	}

	resp, err := s.subledgerClient.CreateTransaction(ctx, &pbSub.CreateTransactionRequest{
		ReferenceId: refID,
		Description: description,
		Entries: []*pbSub.Entry{
			{
				AccountId: fromAccountID,
				Amount:    amount.String(),
				Direction: "DEBIT",
			},
			{
				AccountId: toAccountID,
				Amount:    amount.String(),
				Direction: "CREDIT",
			},
		},
	})

	if err != nil {
		return "", decimal.Zero, fmt.Errorf("failed to create transfer transaction: %w", err)
	}

	newBalance, err := s.repo.GetBalance(ctx, fromAccountID)
	if err != nil {
		return resp.TransactionId, decimal.Zero, err
	}

	s.logger.Infof("Transferred %s from %s to %s, new balance: %s", amount.String(), fromAccountID, toAccountID, newBalance.String())
	return resp.TransactionId, newBalance, nil
}

func (s *Service) GetTransactionHistory(ctx context.Context, accountID string, page, pageSize int) ([]repository.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.GetTransactionHistory(ctx, accountID, page, pageSize)
}
