package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liam-ruiz/budget/internal/auth"
)

// Routes builds and returns the HTTP handler with all routes.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Public auth routes
	r.Post("/api/auth/register", h.authHandler.Register)
	r.Post("/api/auth/login", h.authHandler.Login)

	// Protected routes — require a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(h.JWTSecret))

		// validate route for frontend
		r.Get("/api/auth/validate", h.authHandler.Validate)

		r.Get("/api/accounts", h.acctHandler.GetAccounts)
		r.Get("/api/transactions", h.acctHandler.GetTransactions)
		r.Post("/api/budgets", h.acctHandler.CreateBudget)
		r.Get("/api/budgets", h.acctHandler.GetBudgets)
		r.Post("/api/plaid/exchange", h.plaidHandler.ExchangePlaidPublicToken)
		r.Post("/api/plaid/link-token", h.plaidHandler.CreateLinkToken)
	})

	//
	r.Post("/api/plaid/webhook", h.plaidHandler.HandleWebhook)
	return r
}

// corsMiddleware adds CORS headers so the frontend can reach the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: update to only allow requests from provided environment variable
		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
