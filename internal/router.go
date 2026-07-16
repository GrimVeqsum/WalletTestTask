package internal

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", handler.ChangeBalance)
	mux.HandleFunc("POST /api/v1/wallets", handler.CreateWallet)
	mux.HandleFunc("GET /api/v1/wallets/{id}", handler.GetBalance)
	return mux
}
