package internal

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestRepository(
	t *testing.T,
) (*Repository, *pgxpool.Pool, context.Context) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("integration test skipped: TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		90*time.Second,
	)

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		cancel()
		t.Fatalf("failed to parse test database URL: %v", err)
	}

	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		cancel()
		t.Fatalf("failed to create test database pool: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		cancel()
		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		cancel()
	})

	return NewRepository(db), db, ctx
}

func TestRepositoryConcurrentOperations(t *testing.T) {
	repository, db, ctx := getTestRepository(t)

	walletID, balance, err := repository.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_, _ = db.Exec(
			cleanupCtx,
			"DELETE FROM wallets WHERE id = $1",
			walletID,
		)
	})

	if balance != 0 {
		t.Fatalf("expected initial balance 0, got %d", balance)
	}

	const operationsCount = 1000
	const amount int64 = 1

	depositErrors := make(chan error, operationsCount)

	var depositWaitGroup sync.WaitGroup
	depositWaitGroup.Add(operationsCount)

	for i := 0; i < operationsCount; i++ {
		go func() {
			defer depositWaitGroup.Done()

			_, err := repository.Deposit(
				ctx,
				walletID,
				amount,
			)
			if err != nil {
				depositErrors <- err
			}
		}()
	}

	depositWaitGroup.Wait()
	close(depositErrors)

	for err := range depositErrors {
		t.Fatalf("concurrent deposit failed: %v", err)
	}

	balance, err = repository.GetBalance(ctx, walletID)
	if err != nil {
		t.Fatalf("failed to get balance after deposits: %v", err)
	}

	expectedBalance := int64(operationsCount) * amount

	if balance != expectedBalance {
		t.Fatalf(
			"expected balance %d after deposits, got %d",
			expectedBalance,
			balance,
		)
	}

	withdrawErrors := make(chan error, operationsCount)

	var withdrawWaitGroup sync.WaitGroup
	withdrawWaitGroup.Add(operationsCount)

	for i := 0; i < operationsCount; i++ {
		go func() {
			defer withdrawWaitGroup.Done()

			_, err := repository.Withdraw(
				ctx,
				walletID,
				amount,
			)
			if err != nil {
				withdrawErrors <- err
			}
		}()
	}

	withdrawWaitGroup.Wait()
	close(withdrawErrors)

	for err := range withdrawErrors {
		t.Fatalf("concurrent withdraw failed: %v", err)
	}

	balance, err = repository.GetBalance(ctx, walletID)
	if err != nil {
		t.Fatalf("failed to get balance after withdrawals: %v", err)
	}

	if balance != 0 {
		t.Fatalf("expected final balance 0, got %d", balance)
	}

	_, err = repository.Withdraw(ctx, walletID, 1)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf(
			"expected ErrInsufficientFunds, got %v",
			err,
		)
	}
}

func TestRepositoryWalletNotFound(t *testing.T) {
	repository, _, ctx := getTestRepository(t)

	walletID := uuid.New()

	_, err := repository.GetBalance(ctx, walletID)
	if !errors.Is(err, ErrWalletNotFound) {
		t.Fatalf(
			"expected ErrWalletNotFound from GetBalance, got %v",
			err,
		)
	}

	_, err = repository.Deposit(ctx, walletID, 100)
	if !errors.Is(err, ErrWalletNotFound) {
		t.Fatalf(
			"expected ErrWalletNotFound from Deposit, got %v",
			err,
		)
	}

	_, err = repository.Withdraw(ctx, walletID, 100)
	if !errors.Is(err, ErrWalletNotFound) {
		t.Fatalf(
			"expected ErrWalletNotFound from Withdraw, got %v",
			err,
		)
	}
}
