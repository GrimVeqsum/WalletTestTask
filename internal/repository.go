package internal

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// create
func (r *Repository) CreateWallet(
	ctx context.Context,
) (uuid.UUID, int64, error) {

	query := `INSERT INTO wallets
DEFAULT VALUES
RETURNING id,balance
`
	var id uuid.UUID
	var balance int64
	err := r.db.QueryRow(ctx, query).Scan(&id, &balance)
	if err != nil {
		return uuid.Nil, 0, err
	}
	return id, balance, nil
}

// deposit
func (r *Repository) Deposit(
	ctx context.Context,
	walletID uuid.UUID,
	amount int64,
) (int64, error) {
	query := `UPDATE wallets
SET balance = balance + $2
WHERE id = $1
RETURNING balance;`
	var balance int64
	err := r.db.QueryRow(
		ctx,
		query,
		walletID,
		amount,
	).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}

	if err != nil {
		return 0, err
	}
	return balance, nil
}

// withdraw
func (r *Repository) Withdraw(
	ctx context.Context,
	walletID uuid.UUID,
	amount int64,
) (int64, error) {
	query := `UPDATE wallets
SET balance = balance - $2
WHERE id = $1
  AND balance >= $2
RETURNING balance;
	`
	var balance int64
	err := r.db.QueryRow(
		ctx,
		query,
		walletID,
		amount,
	).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		query = `SELECT balance
FROM wallets
WHERE id = $1
`
		err = r.db.QueryRow(
			ctx,
			query,
			walletID,
		).Scan(&balance)

		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrWalletNotFound
		}

		if err != nil {
			return 0, err
		}

		return 0, ErrInsufficientFunds
	}

	if err != nil {
		return 0, err
	}
	return balance, nil
}

// get balance
func (r *Repository) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	query := `SELECT balance
FROM wallets
WHERE id = $1
`
	var balance int64
	err := r.db.QueryRow(
		ctx,
		query,
		walletID,
	).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}
