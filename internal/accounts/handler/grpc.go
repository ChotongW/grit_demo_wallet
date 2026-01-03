package handler

import (
	"context"
	"errors"
	"math"

	accountErrors "github.com/ChotongW/grit_demo_wallet/internal/accounts/errors"
	"github.com/ChotongW/grit_demo_wallet/internal/accounts/service"
	pb "github.com/ChotongW/grit_demo_wallet/pb/accounts"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedAccountsServiceServer
	service *service.Service
	logger  *logrus.Entry
}

func NewGRPCHandler(svc *service.Service, logger *logrus.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: svc,
		logger: logger.WithFields(logrus.Fields{
			"package": "accounts/handler",
		}),
	}
}

func (h *GRPCHandler) loggerWithRequestID(ctx context.Context) *logrus.Entry {
	reqID := requestid.FromContext(ctx)
	if reqID != "" {
		return h.logger.WithField("request_id", reqID)
	}
	return h.logger
}

func (h *GRPCHandler) mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, accountErrors.ErrAccountNotFound) {
		return status.Errorf(codes.NotFound, "%v", err)
	}
	if errors.Is(err, accountErrors.ErrEmailAlreadyExists) {
		return status.Errorf(codes.AlreadyExists, "%v", err)
	}
	if errors.Is(err, accountErrors.ErrInsufficientBalance) {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	if errors.Is(err, accountErrors.ErrInvalidReferrer) ||
		errors.Is(err, accountErrors.ErrTransferToSameAccount) ||
		errors.Is(err, accountErrors.ErrDepositAmountMustBePositive) ||
		errors.Is(err, accountErrors.ErrWithdrawAmountMustBePositive) ||
		errors.Is(err, accountErrors.ErrTransferAmountMustBePositive) {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return status.Errorf(codes.Internal, "%v", err)
}

func (h *GRPCHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	var initialBalance decimal.Decimal
	if req.InitialBalance != "" {
		var err error
		initialBalance, err = decimal.NewFromString(req.InitialBalance)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid initial balance: %v", err)
		}
		if initialBalance.LessThan(decimal.Zero) {
			return nil, status.Errorf(codes.InvalidArgument, "initial balance cannot be negative")
		}
	}

	account, err := h.service.CreateAccount(ctx, req.Email, initialBalance, req.ReferrerAccountId)
	if err != nil {
		logger.Errorf("failed to create account: %v", err)
		return nil, h.mapError(err)
	}

	balance, err := h.service.GetBalance(ctx, account.AccountID)
	if err != nil {
		logger.Warnf("failed to get balance for new account: %v", err)
		balance = decimal.Zero
	}

	logger.Infof("created account: %s", account.AccountID)
	return &pb.CreateAccountResponse{
		Success:   true,
		AccountId: account.AccountID,
		Message:   "Account created successfully",
		Account: &pb.Account{
			AccountId:         account.AccountID,
			AccountType:       account.AccountType,
			UserId:            account.UserID,
			Email:             account.Email,
			ReferrerAccountId: account.ReferrerAccountID,
			Balance:           balance.String(),
			CreatedAt:         account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (h *GRPCHandler) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	account, balance, err := h.service.GetAccount(ctx, req.AccountId)
	if err != nil {
		logger.Errorf("account not found: %v", err)
		return nil, h.mapError(err)
	}

	logger.Infof("retrieved account: %s", req.AccountId)
	return &pb.GetAccountResponse{
		Account: &pb.Account{
			AccountId:         account.AccountID,
			AccountType:       account.AccountType,
			UserId:            account.UserID,
			Email:             account.Email,
			ReferrerAccountId: account.ReferrerAccountID,
			Balance:           balance.String(),
			CreatedAt:         account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (h *GRPCHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	balance, err := h.service.GetBalance(ctx, req.AccountId)
	if err != nil {
		logger.Errorf("failed to get balance: %v", err)
		return nil, h.mapError(err)
	}

	logger.Infof("retrieved balance for account: %s", req.AccountId)
	return &pb.GetBalanceResponse{
		AccountId: req.AccountId,
		Balance:   balance.String(),
		Currency:  "USD",
	}, nil
}

func (h *GRPCHandler) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
	}

	txnID, newBalance, err := h.service.Deposit(ctx, req.AccountId, amount, req.Description)
	if err != nil {
		logger.Errorf("failed to deposit: %v", err)
		return nil, h.mapError(err)
	}

	logger.Infof("deposit successful: account=%s, amount=%s, txn=%s", req.AccountId, req.Amount, txnID)
	return &pb.DepositResponse{
		Success:       true,
		TransactionId: txnID,
		NewBalance:    newBalance.String(),
		Message:       "Deposit successful",
	}, nil
}

func (h *GRPCHandler) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
	}

	txnID, newBalance, err := h.service.Withdraw(ctx, req.AccountId, amount, req.Description)
	if err != nil {
		logger.Errorf("failed to withdraw: %v", err)
		return nil, h.mapError(err)
	}

	logger.Infof("withdrawal successful: account=%s, amount=%s, txn=%s", req.AccountId, req.Amount, txnID)
	return &pb.WithdrawResponse{
		Success:       true,
		TransactionId: txnID,
		NewBalance:    newBalance.String(),
		Message:       "Withdrawal successful",
	}, nil
}

func (h *GRPCHandler) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
	}

	txnID, newBalance, err := h.service.Transfer(ctx, req.FromAccountId, req.ToAccountId, amount, req.Description)
	if err != nil {
		logger.Errorf("failed to transfer: %v", err)
		return nil, h.mapError(err)
	}

	logger.Infof("transfer successful: from=%s, to=%s, amount=%s, txn=%s", req.FromAccountId, req.ToAccountId, req.Amount, txnID)
	return &pb.TransferResponse{
		Success:       true,
		TransactionId: txnID,
		NewBalance:    newBalance.String(),
		Message:       "Transfer successful",
	}, nil
}

func (h *GRPCHandler) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	transactions, totalCount, err := h.service.GetTransactionHistory(ctx, req.AccountId, page, pageSize)
	if err != nil {
		logger.Errorf("failed to get transaction history: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get transaction history: %v", err)
	}

	// Convert to proto transactions
	protoTransactions := make([]*pb.Transaction, len(transactions))
	for i, txn := range transactions {
		protoTransactions[i] = &pb.Transaction{
			Id:            txn.ID,
			TransactionId: txn.TransactionID,
			AccountId:     txn.AccountID,
			Amount:        txn.Amount.String(),
			Direction:     txn.Direction,
			ReferenceId:   txn.ReferenceID,
			Description:   txn.Description,
			CreatedAt:     txn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	logger.Infof("retrieved transaction history: account=%s, count=%d", req.AccountId, len(transactions))
	return &pb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(totalCount),
		Page:         int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
	}, nil
}
