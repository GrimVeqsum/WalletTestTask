package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service WalletService
}

type WalletService interface {
	ChangeBalance(
		ctx context.Context,
		walletID uuid.UUID,
		operation OperationType,
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

func NewWalletHandler(service WalletService) *Handler {
	return &Handler{
		service: service,
	}
}

// create
func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {

	walletID, balance, err := h.service.CreateWallet(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	response := WalletResponse{
		WalletID: walletID,
		Balance:  balance,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// change balance
func (h *Handler) ChangeBalance(w http.ResponseWriter, r *http.Request) {
	var cbr ChangeBalanceRequest
	err := json.NewDecoder(r.Body).Decode(&cbr)
	if err != nil {
		http.Error(w, "invalid_json", 400)
		return
	}
	balance, err := h.service.ChangeBalance(
		r.Context(),
		cbr.WalletID,
		cbr.OperationType,
		cbr.Amount,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount):
			http.Error(w, ErrInvalidAmount.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrInvalidOperation):
			http.Error(w, ErrInvalidOperation.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrWalletNotFound):
			http.Error(w, ErrWalletNotFound.Error(), http.StatusNotFound)
		case errors.Is(err, ErrInsufficientFunds):
			http.Error(w, ErrInsufficientFunds.Error(), http.StatusConflict)
		default:
			slog.Error("failed to change wallet balance", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	response := WalletResponse{
		Balance:  balance,
		WalletID: cbr.WalletID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// get balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid wallet id", http.StatusBadRequest)
		return
	}
	balance, err := h.service.GetBalance(r.Context(), ID)
	if err != nil {
		if errors.Is(err, ErrWalletNotFound) {
			http.Error(w, ErrWalletNotFound.Error(), http.StatusNotFound)
		} else {
			slog.Error("failed to get wallet balance", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	response := WalletResponse{
		Balance:  balance,
		WalletID: ID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
