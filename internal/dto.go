package internal

import "github.com/google/uuid"

type ChangeBalanceRequest struct {
	WalletID      uuid.UUID     `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64         `json:"amount"`
}

type WalletResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}
