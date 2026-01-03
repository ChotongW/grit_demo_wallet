package service

import (
	"context"
	"fmt"

	"github.com/ChotongW/grit_demo_wallet/internal/subledger/repository"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var (
	DEBIT  = "DEBIT"
	CREDIT = "CREDIT"
)

type Service struct {
	repo   *repository.Repository
	logger logrus.FieldLogger
}

func NewService(repo *repository.Repository, logger *logrus.Logger) *Service {
	newLogger := logger.WithFields(
		logrus.Fields{
			"package": "service",
		},
	)
	return &Service{
		repo:   repo,
		logger: newLogger,
	}
}

func (s *Service) CreateTransaction(ctx context.Context, refID string, desc string, entries []repository.TransactionEntry) error {
	if len(entries) < 2 {
		s.logger.Errorf("at least 2 entries required for double-entry accounting")
		return fmt.Errorf("at least 2 entries required for double-entry accounting")
	}
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero

	for _, entry := range entries {
		if entry.Direction == DEBIT {
			totalDebits = totalDebits.Add(entry.Amount)
		} else if entry.Direction == CREDIT {
			totalCredits = totalCredits.Add(entry.Amount)
		} else {
			return fmt.Errorf("invalid direction: %s", entry.Direction)
		}
	}

	if !totalDebits.Equal(totalCredits) {
		return fmt.Errorf("debits (%s) must equal credits (%s)", totalDebits.String(), totalCredits.String())
	}

	return s.repo.CreateTransaction(ctx, refID, desc, entries)
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	return s.repo.GetBalance(ctx, accountID)
}
