package api

import (
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budget/internal/api/types"
	"github.com/liam-ruiz/budget/internal/auth"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/dependencies"
)

type AccountHandler struct {
	container *dependencies.Container
}

func NewAccountHandler(container *dependencies.Container) *AccountHandler {
	return &AccountHandler{
		container: container,
	}
}

// GetAccounts returns all linked bank accounts for the authenticated user.
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetAccounts] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.container.AccountSvc.GetAccounts(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting accounts: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch accounts")
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

// GetTransactions returns all transactions for the authenticated user.
func (h *AccountHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetTransactions] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	txns, err := h.container.TransactionSvc.GetByUser(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting transactions: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	writeJSON(w, http.StatusOK, txns)
}

// CreateBudget creates a new budget for the authenticated user.
func (h *AccountHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CreateBudget] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req types.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Category == "" || req.LimitAmount == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "category, limit_amount, and start_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		log.Printf("Error parsing start date: %v\n", err)
		writeError(w, http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
		return
	}

	period := req.Period
	if period == "" {
		period = "monthly"
	}
	// parse limit amount ("10.00" -> 1000 to be used for Numeric)
	combined := strings.ReplaceAll(req.LimitAmount, ".", "")
	limitAmount, err := strconv.Atoi(combined)
	if err != nil {
		log.Printf("Error parsing limit amount: %v\n", err)
		writeError(w, http.StatusBadRequest, "limit_amount must be a valid number")
		return
	}

	params := sqlcdb.CreateBudgetParams{
		AppUserID:    userID,
		Category:     req.Category,
		LimitAmount:  pgtype.Numeric{Valid: true, Int: big.NewInt(int64(limitAmount))},
		BudgetPeriod: period,
		StartDate:    pgtype.Date{Time: startDate, Valid: true},
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			log.Printf("Error parsing end date: %v\n", err)
			writeError(w, http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
			return
		}
		params.EndDate = pgtype.Date{Time: endDate, Valid: true}
	}

	budget, err := h.container.BudgetSvc.CreateBudget(r.Context(), params)
	if err != nil {
		log.Printf("Error creating budget: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to create budget")
		return
	}

	writeJSON(w, http.StatusCreated, budget)
}

// GetBudgets returns all budgets for the authenticated user.
func (h *AccountHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetBudgets] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.container.BudgetSvc.GetBudgets(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting budgets: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch budgets")
		return
	}

	writeJSON(w, http.StatusOK, list)
}