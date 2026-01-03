package repository

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	accountErrors "github.com/ChotongW/grit_demo_wallet/internal/accounts/errors"
)

type Account struct {
	AccountID         string
	AccountType       string
	UserID            *string
	Email             *string
	ReferrerAccountID *string
	CreatedAt         time.Time
}

type Transaction struct {
	ID            string
	TransactionID string
	AccountID     string
	Amount        decimal.Decimal
	Direction     string
	ReferenceID   string
	Description   string
	CreatedAt     time.Time
}

type Repository struct {
	pool   *pgxpool.Pool
	logger *logrus.Entry
}

func NewRepository(pool *pgxpool.Pool, logger *logrus.Logger) *Repository {
	return &Repository{
		pool: pool,
		logger: logger.WithFields(logrus.Fields{
			"package": "accounts/repository",
		}),
	}
}

func (r *Repository) CreateAccount(ctx context.Context, email, referrerAccountID string) (*Account, error) {
	accountID := uuid.New().String()

	query := `
		INSERT INTO accounts (account_id, account_type, user_id, email, referrer_account_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING account_id, account_type, user_id, email, referrer_account_id, created_at
	`

	var account Account
	err := r.pool.QueryRow(ctx, query,
		accountID,
		"USER",
		accountID,
		email,
		referrerAccountID,
	).Scan(
		&account.AccountID,
		&account.AccountType,
		&account.UserID,
		&account.Email,
		&account.ReferrerAccountID,
		&account.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: %v", accountErrors.ErrEmailAlreadyExists, err)
		}
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	r.logger.Infof("Created account: %s for email: %s", accountID, email)
	return &account, nil
}

func (r *Repository) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	query := `
		SELECT account_id, account_type, user_id, email, referrer_account_id, created_at
		FROM accounts
		WHERE account_id = $1
	`

	var account Account
	err := r.pool.QueryRow(ctx, query, accountID).Scan(
		&account.AccountID,
		&account.AccountType,
		&account.UserID,
		&account.Email,
		&account.ReferrerAccountID,
		&account.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: account %s", accountErrors.ErrAccountNotFound, accountID)
		}
		return nil, fmt.Errorf("failed to get account %s: %w", accountID, err)
	}

	return &account, nil
}

func (r *Repository) GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	query := `SELECT COALESCE(amount, 0) FROM balances WHERE account_id = $1`

	var balance decimal.Decimal
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&balance)
	if err != nil {
		return decimal.Zero, nil
	}

	return balance, nil
}

func (r *Repository) AccountExists(ctx context.Context, accountID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) GetTransactionHistory(ctx context.Context, accountID string, page, pageSize int) ([]Transaction, int, error) {
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM ledger_entries WHERE account_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, accountID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transaction count: %w", err)
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT id, transaction_id, account_id, amount, direction,
		       COALESCE(reference_id, '') as reference_id,
		       COALESCE(description, '') as description,
		       created_at
		FROM ledger_entries
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, accountID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var txn Transaction
		err := rows.Scan(
			&txn.ID,
			&txn.TransactionID,
			&txn.AccountID,
			&txn.Amount,
			&txn.Direction,
			&txn.ReferenceID,
			&txn.Description,
			&txn.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	return transactions, totalCount, nil
}
