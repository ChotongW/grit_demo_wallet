package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type TransactionEntry struct {
	AccountID string
	Amount    decimal.Decimal
	Direction string
}

type Repository struct {
	pool   *pgxpool.Pool
	logger logrus.FieldLogger
}

func NewRepository(pool *pgxpool.Pool, logger *logrus.Logger) *Repository {
	newLogger := logger.WithFields(
		logrus.Fields{
			"package": "service",
		},
	)
	return &Repository{
		pool:   pool,
		logger: newLogger,
	}
}

func (r *Repository) CreateTransaction(ctx context.Context, refID string, desc string, entries []TransactionEntry) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	if len(entries) == 0 {
		return nil
	}

	trxID := uuid.New().String()
	timestamp := time.Now()

	var ledgerArgs []interface{}

	balanceMap := make(map[string]decimal.Decimal)

	for _, entry := range entries {
		entryID := uuid.New().String()

		ledgerArgs = append(ledgerArgs,
			entryID,
			trxID,
			entry.AccountID,
			entry.Amount,
			entry.Direction,
			refID,
			desc,
			timestamp,
		)

		amount := entry.Amount
		if entry.Direction == "DEBIT" {
			amount = amount.Neg()
		}

		if val, ok := balanceMap[entry.AccountID]; ok {
			balanceMap[entry.AccountID] = val.Add(amount)
		} else {
			balanceMap[entry.AccountID] = amount
		}
	}

	placeholders := buildPlaceholders(1, len(entries), 8)
	queryLedger := fmt.Sprintf(`
			INSERT INTO ledger_entries (id, transaction_id, account_id, amount, direction, reference_id, description, created_at)
			VALUES %s
		`, placeholders)

	_, err = tx.Exec(ctx, queryLedger, ledgerArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entries: %w", err)
	}

	accountIDs := make([]string, 0, len(balanceMap))
	for accID := range balanceMap {
		accountIDs = append(accountIDs, accID)
	}
	sort.Strings(accountIDs)

	var balanceArgs []interface{}
	for _, accID := range accountIDs {
		amount := balanceMap[accID]
		balanceArgs = append(balanceArgs,
			accID,
			amount,
			timestamp,
		)
	}

	balancePlaceholders := buildPlaceholders(1, len(accountIDs), 3)
	queryBalance := fmt.Sprintf(`
			INSERT INTO balances (account_id, amount, updated_at)
			VALUES %s
			ON CONFLICT (account_id)
			DO UPDATE SET 
				amount = balances.amount + EXCLUDED.amount,
				updated_at = EXCLUDED.updated_at
		`, balancePlaceholders)

	_, err = tx.Exec(ctx, queryBalance, balanceArgs...)
	if err != nil {
		return fmt.Errorf("failed to update balances: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func buildPlaceholders(startCount, rows, cols int) string {
	var b strings.Builder
	for i := 0; i < rows; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("(")
		for j := 0; j < cols; j++ {
			if j > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, "$%d", startCount+j)
		}
		b.WriteString(")")
		startCount += cols
	}
	return b.String()
}

func (r *Repository) GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	query := `SELECT amount FROM balances WHERE account_id = $1`

	var amount decimal.Decimal
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&amount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance for account %s: %w", accountID, err)
	}

	return amount, nil
}
