package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type handlerServiceStub struct {
	changeBalanceFn func(
		context.Context,
		uuid.UUID,
		OperationType,
		int64,
	) (int64, error)

	getBalanceFn func(
		context.Context,
		uuid.UUID,
	) (int64, error)

	createWalletFn func(
		context.Context,
	) (uuid.UUID, int64, error)
}

func (s *handlerServiceStub) ChangeBalance(
	ctx context.Context,
	walletID uuid.UUID,
	operation OperationType,
	amount int64,
) (int64, error) {
	if s.changeBalanceFn == nil {
		return 0, nil
	}

	return s.changeBalanceFn(
		ctx,
		walletID,
		operation,
		amount,
	)
}

func (s *handlerServiceStub) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	if s.getBalanceFn == nil {
		return 0, nil
	}

	return s.getBalanceFn(ctx, walletID)
}

func (s *handlerServiceStub) CreateWallet(
	ctx context.Context,
) (uuid.UUID, int64, error) {
	if s.createWalletFn == nil {
		return uuid.Nil, 0, nil
	}

	return s.createWalletFn(ctx)
}

func TestHandlerCreateWallet(t *testing.T) {
	expectedWalletID := uuid.New()

	service := &handlerServiceStub{
		createWalletFn: func(
			context.Context,
		) (uuid.UUID, int64, error) {
			return expectedWalletID, 0, nil
		},
	}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/wallets",
		nil,
	)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			responseRecorder.Code,
		)
	}

	var response WalletResponse

	err := json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.WalletID != expectedWalletID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			expectedWalletID,
			response.WalletID,
		)
	}

	if response.Balance != 0 {
		t.Fatalf("expected balance 0, got %d", response.Balance)
	}
}

func TestHandlerCreateWalletInternalError(t *testing.T) {
	service := &handlerServiceStub{
		createWalletFn: func(
			context.Context,
		) (uuid.UUID, int64, error) {
			return uuid.Nil, 0, errors.New("database error")
		},
	}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/wallets",
		nil,
	)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			responseRecorder.Code,
		)
	}
}

func TestHandlerChangeBalance(t *testing.T) {
	walletID := uuid.New()

	service := &handlerServiceStub{
		changeBalanceFn: func(
			ctx context.Context,
			receivedWalletID uuid.UUID,
			operation OperationType,
			amount int64,
		) (int64, error) {
			if receivedWalletID != walletID {
				t.Fatalf(
					"expected wallet ID %s, got %s",
					walletID,
					receivedWalletID,
				)
			}

			if operation != OperationDeposit {
				t.Fatalf(
					"expected operation %s, got %s",
					OperationDeposit,
					operation,
				)
			}

			if amount != 1000 {
				t.Fatalf("expected amount 1000, got %d", amount)
			}

			return 1000, nil
		},
	}

	requestBody, err := json.Marshal(ChangeBalanceRequest{
		WalletID:      walletID,
		OperationType: OperationDeposit,
		Amount:        1000,
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/wallet",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			responseRecorder.Code,
		)
	}

	var response WalletResponse

	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.WalletID != walletID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			walletID,
			response.WalletID,
		)
	}

	if response.Balance != 1000 {
		t.Fatalf(
			"expected balance 1000, got %d",
			response.Balance,
		)
	}
}

func TestHandlerChangeBalanceInvalidJSON(t *testing.T) {
	service := &handlerServiceStub{}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/wallet",
		bytes.NewBufferString("{invalid"),
	)
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			responseRecorder.Code,
		)
	}
}

func TestHandlerChangeBalanceErrors(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "invalid amount",
			serviceError:   ErrInvalidAmount,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid operation",
			serviceError:   ErrInvalidOperation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wallet not found",
			serviceError:   ErrWalletNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "insufficient funds",
			serviceError:   ErrInsufficientFunds,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "internal error",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &handlerServiceStub{
				changeBalanceFn: func(
					context.Context,
					uuid.UUID,
					OperationType,
					int64,
				) (int64, error) {
					return 0, test.serviceError
				},
			}

			requestBody, err := json.Marshal(ChangeBalanceRequest{
				WalletID:      uuid.New(),
				OperationType: OperationDeposit,
				Amount:        100,
			})
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			handler := NewWalletHandler(service)
			router := NewRouter(handler)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/wallet",
				bytes.NewReader(requestBody),
			)
			request.Header.Set(
				"Content-Type",
				"application/json",
			)

			responseRecorder := httptest.NewRecorder()

			router.ServeHTTP(responseRecorder, request)

			if responseRecorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					test.expectedStatus,
					responseRecorder.Code,
				)
			}
		})
	}
}

func TestHandlerGetBalance(t *testing.T) {
	walletID := uuid.New()

	service := &handlerServiceStub{
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

			return 700, nil
		},
	}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/wallets/"+walletID.String(),
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			responseRecorder.Code,
		)
	}

	var response WalletResponse

	err := json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.WalletID != walletID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			walletID,
			response.WalletID,
		)
	}

	if response.Balance != 700 {
		t.Fatalf(
			"expected balance 700, got %d",
			response.Balance,
		)
	}
}

func TestHandlerGetBalanceInvalidUUID(t *testing.T) {
	service := &handlerServiceStub{}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/wallets/invalid-uuid",
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			responseRecorder.Code,
		)
	}
}

func TestHandlerGetBalanceNotFound(t *testing.T) {
	service := &handlerServiceStub{
		getBalanceFn: func(
			context.Context,
			uuid.UUID,
		) (int64, error) {
			return 0, ErrWalletNotFound
		},
	}

	handler := NewWalletHandler(service)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/wallets/"+uuid.New().String(),
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			responseRecorder.Code,
		)
	}
}
