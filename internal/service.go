package internal

import (
	"context"

	"github.com/google/uuid"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type WalletRepository interface {
	Deposit(
		ctx context.Context,
		walletID uuid.UUID,
		amount int64,
	) (int64, error)

	Withdraw(
		ctx context.Context,
		walletID uuid.UUID,
		amount int64,
	) (int64, error)

	GetBalance(
		ctx context.Context,
		walletID uuid.UUID,
	) (int64, error)

	CreateWallet(
		ctx context.Context,
	) (uuid.UUID, int64, error)
}

type Service struct {
	repository WalletRepository
}

func NewService(repository WalletRepository) *Service {
	return &Service{
		repository: repository,
	}
}

// get
func (s *Service) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	return s.repository.GetBalance(ctx, walletID)
}

// create
func (s *Service) CreateWallet(
	ctx context.Context,
) (uuid.UUID, int64, error) {
	return s.repository.CreateWallet(ctx)
}

// change
func (s *Service) ChangeBalance(
	ctx context.Context,
	walletID uuid.UUID,
	operation OperationType,
	amount int64,
) (int64, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	switch operation {
	case OperationDeposit:
		return s.repository.Deposit(ctx, walletID, amount)
	case OperationWithdraw:
		return s.repository.Withdraw(ctx, walletID, amount)
	default:
		return 0, ErrInvalidOperation
	}

}
