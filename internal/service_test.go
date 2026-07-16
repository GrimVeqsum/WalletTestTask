package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type serviceRepositoryStub struct {
	depositFn      func(context.Context, uuid.UUID, int64) (int64, error)
	withdrawFn     func(context.Context, uuid.UUID, int64) (int64, error)
	getBalanceFn   func(context.Context, uuid.UUID) (int64, error)
	createWalletFn func(context.Context) (uuid.UUID, int64, error)
}

func (s *serviceRepositoryStub) Deposit(
	ctx context.Context,
	walletID uuid.UUID,
	amount int64,
) (int64, error) {
	if s.depositFn == nil {
		return 0, nil
	}

	return s.depositFn(ctx, walletID, amount)
}

func (s *serviceRepositoryStub) Withdraw(
	ctx context.Context,
	walletID uuid.UUID,
	amount int64,
) (int64, error) {
	if s.withdrawFn == nil {
		return 0, nil
	}

	return s.withdrawFn(ctx, walletID, amount)
}

func (s *serviceRepositoryStub) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	if s.getBalanceFn == nil {
		return 0, nil
	}

	return s.getBalanceFn(ctx, walletID)
}

func (s *serviceRepositoryStub) CreateWallet(
	ctx context.Context,
) (uuid.UUID, int64, error) {
	if s.createWalletFn == nil {
		return uuid.Nil, 0, nil
	}

	return s.createWalletFn(ctx)
}

func TestServiceChangeBalanceInvalidAmount(t *testing.T) {
	repository := &serviceRepositoryStub{
		depositFn: func(
			context.Context,
			uuid.UUID,
			int64,
		) (int64, error) {
			t.Fatal("repository must not be called")
			return 0, nil
		},
		withdrawFn: func(
			context.Context,
			uuid.UUID,
			int64,
		) (int64, error) {
			t.Fatal("repository must not be called")
			return 0, nil
		},
	}

	service := NewService(repository)
	walletID := uuid.New()

	amounts := []int64{0, -1, -100}

	for _, amount := range amounts {
		_, err := service.ChangeBalance(
			context.Background(),
			walletID,
			OperationDeposit,
			amount,
		)

		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf(
				"amount %d: expected ErrInvalidAmount, got %v",
				amount,
				err,
			)
		}
	}
}

func TestServiceChangeBalanceInvalidOperation(t *testing.T) {
	repository := &serviceRepositoryStub{}
	service := NewService(repository)

	_, err := service.ChangeBalance(
		context.Background(),
		uuid.New(),
		OperationType("TRANSFER"),
		100,
	)

	if !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("expected ErrInvalidOperation, got %v", err)
	}
}

func TestServiceChangeBalanceDeposit(t *testing.T) {
	walletID := uuid.New()
	expectedBalance := int64(1500)

	repository := &serviceRepositoryStub{
		depositFn: func(
			ctx context.Context,
			receivedWalletID uuid.UUID,
			amount int64,
		) (int64, error) {
			if receivedWalletID != walletID {
				t.Fatalf(
					"expected wallet ID %s, got %s",
					walletID,
					receivedWalletID,
				)
			}

			if amount != 500 {
				t.Fatalf("expected amount 500, got %d", amount)
			}

			return expectedBalance, nil
		},
	}

	service := NewService(repository)

	balance, err := service.ChangeBalance(
		context.Background(),
		walletID,
		OperationDeposit,
		500,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if balance != expectedBalance {
		t.Fatalf(
			"expected balance %d, got %d",
			expectedBalance,
			balance,
		)
	}
}

func TestServiceChangeBalanceWithdraw(t *testing.T) {
	walletID := uuid.New()
	expectedBalance := int64(600)

	repository := &serviceRepositoryStub{
		withdrawFn: func(
			ctx context.Context,
			receivedWalletID uuid.UUID,
			amount int64,
		) (int64, error) {
			if receivedWalletID != walletID {
				t.Fatalf(
					"expected wallet ID %s, got %s",
					walletID,
					receivedWalletID,
				)
			}

			if amount != 400 {
				t.Fatalf("expected amount 400, got %d", amount)
			}

			return expectedBalance, nil
		},
	}

	service := NewService(repository)

	balance, err := service.ChangeBalance(
		context.Background(),
		walletID,
		OperationWithdraw,
		400,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if balance != expectedBalance {
		t.Fatalf(
			"expected balance %d, got %d",
			expectedBalance,
			balance,
		)
	}
}

func TestServiceReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repository := &serviceRepositoryStub{
		depositFn: func(
			context.Context,
			uuid.UUID,
			int64,
		) (int64, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repository)

	_, err := service.ChangeBalance(
		context.Background(),
		uuid.New(),
		OperationDeposit,
		100,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceGetBalance(t *testing.T) {
	walletID := uuid.New()
	expectedBalance := int64(900)

	repository := &serviceRepositoryStub{
		getBalanceFn: func(
			ctx context.Context,
			receivedWalletID uuid.UUID,
		) (int64, error) {
			if receivedWalletID != walletID {
				t.Fatalf(
					"expected wallet ID %s, got %s",
					walletID,
					receivedWalletID,
				)
			}

			return expectedBalance, nil
		},
	}

	service := NewService(repository)

	balance, err := service.GetBalance(
		context.Background(),
		walletID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if balance != expectedBalance {
		t.Fatalf(
			"expected balance %d, got %d",
			expectedBalance,
			balance,
		)
	}
}

func TestServiceCreateWallet(t *testing.T) {
	expectedWalletID := uuid.New()
	expectedBalance := int64(0)

	repository := &serviceRepositoryStub{
		createWalletFn: func(
			context.Context,
		) (uuid.UUID, int64, error) {
			return expectedWalletID, expectedBalance, nil
		},
	}

	service := NewService(repository)

	walletID, balance, err := service.CreateWallet(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if walletID != expectedWalletID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			expectedWalletID,
			walletID,
		)
	}

	if balance != expectedBalance {
		t.Fatalf(
			"expected balance %d, got %d",
			expectedBalance,
			balance,
		)
	}
}
