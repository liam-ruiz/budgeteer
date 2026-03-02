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
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	// Protected routes — require a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(h.container.Cfg.JWTSecret))

		r.Get("/accounts", h.GetAccounts)
		r.Get("/transactions", h.GetTransactions)
		r.Post("/budgets", h.CreateBudget)
		r.Get("/budgets", h.GetBudgets)
		r.Post("/plaid/exchange", h.ExchangePlaidPublicToken)
		r.Post("/plaid/link-token", h.CreateLinkToken)
	})

	r.Post("/plaid/webhook", h.HandleWebhook)
	return r
}

// corsMiddleware adds CORS headers so the frontend can reach the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
