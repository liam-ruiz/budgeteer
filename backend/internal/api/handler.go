package api

import (
	"encoding/json"
	"net/http"

	"github.com/liam-ruiz/budget/internal/api/types"
	"github.com/liam-ruiz/budget/internal/dependencies"
)

// Handler holds all service dependencies for the API.
type Handler struct {
	acctHandler *AccountHandler
	plaidHandler *PlaidHandler
	authHandler *AuthHandler
	JWTSecret string
}


// NewHandler creates a new API handler with all service dependencies.
func NewHandler(
	container *dependencies.Container,
) *Handler {
	return &Handler{
		acctHandler: NewAccountHandler(container),
		plaidHandler: NewPlaidHandler(container),
		authHandler: NewAuthHandler(container),
		JWTSecret: container.Cfg.JWTSecret,
	}
}


// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, types.ErrorResponse{Error: msg})
}
