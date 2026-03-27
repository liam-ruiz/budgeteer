package api

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budgeteer/internal/api/types"
	"github.com/liam-ruiz/budgeteer/internal/auth"
	"github.com/liam-ruiz/budgeteer/internal/bank_accounts"
	"github.com/liam-ruiz/budgeteer/internal/budgets"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
	"github.com/liam-ruiz/budgeteer/internal/dependencies"
	"github.com/liam-ruiz/budgeteer/internal/transactions"
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

// GetAccount returns a single linked bank account for the authenticated user.
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetAccount] %s %s", r.Method, r.URL.Path)
	// TODO: verify account belongs to user
	_, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	account, err := h.container.AccountSvc.GetAccount(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, bank_accounts.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}

		log.Printf("Error getting account: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch account")
		return
	}

	writeJSON(w, http.StatusOK, account)
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

// GetTransactionsByAccount returns all transactions for a single linked account.
func (h *AccountHandler) GetTransactionsByAccount(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetTransactionsByAccount] %s %s", r.Method, r.URL.Path)
	// TODO: verify account belongs to user
	_, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	// TODO: see if this breaks the frontend response by removing account name from the transaction response
	txns, err := h.container.TransactionSvc.GetByAccount(r.Context(), accountID)
	if err != nil {
		log.Printf("Error getting account transactions: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch account transactions")
		return
	}

	writeJSON(w, http.StatusOK, txns)
}

// DeleteAccount removes a linked bank account for the authenticated user.
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DeleteAccount] %s %s", r.Method, r.URL.Path)

	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	err = h.container.AccountSvc.DeleteAccount(r.Context(), userID, accountID)
	if err != nil {
		if errors.Is(err, bank_accounts.ErrAccountNotFound) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}

		log.Printf("Error deleting account: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// DeleteTransaction removes a transaction for the authenticated user.
func (h *AccountHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DeleteTransaction] %s %s", r.Method, r.URL.Path)

	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	transactionID := chi.URLParam(r, "id")
	if transactionID == "" {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	err = h.container.TransactionSvc.DeleteTransaction(r.Context(), userID, transactionID)
	if err != nil {
		if errors.Is(err, transactions.ErrTransactionNotFound) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}

		log.Printf("Error deleting transaction: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AccountHandler) UpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	log.Printf("[UpdateTransactionCategory] %s %s", r.Method, r.URL.Path)

	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	transactionID := chi.URLParam(r, "id")
	if transactionID == "" {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var req types.UpdateTransactionCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Category) == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}

	err = h.container.TransactionSvc.UpdateCategory(r.Context(), userID, transactionID, req.Category)
	if err != nil {
		switch {
		case errors.Is(err, transactions.ErrTransactionNotFound):
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		case errors.Is(err, transactions.ErrInvalidTransactionCategory):
			writeError(w, http.StatusBadRequest, "invalid transaction category")
			return
		default:
			log.Printf("Error updating transaction category: %v\n", err)
			writeError(w, http.StatusInternalServerError, "failed to update transaction category")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
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
	if req.Name == "" || req.LimitAmount == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "name, limit_amount, and start_date are required")
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
	combined := strings.ReplaceAll(req.LimitAmount, ".", "")
	limitAmount, err := strconv.Atoi(combined)
	if err != nil {
		log.Printf("Error parsing limit amount: %v\n", err)
		writeError(w, http.StatusBadRequest, "limit_amount must be a valid number")
		return
	}

	var category pgtype.Text
	if req.Category != nil && *req.Category != "" {
		category = pgtype.Text{String: *req.Category, Valid: true}
	}

	params := sqlcdb.CreateBudgetParams{
		AppUserID:    userID,
		Name:         req.Name,
		Category:     category,
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

// GetBudget returns a single budget for the authenticated user.
func (h *AccountHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetBudget] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	budget, err := h.container.BudgetSvc.GetBudget(r.Context(), userID, budgetID)
	if err != nil {
		if errors.Is(err, budgets.ErrBudgetNotFound) {
			writeError(w, http.StatusNotFound, "budget not found")
			return
		}

		log.Printf("Error getting budget: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch budget")
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

// GetTransactionsByBudget returns all transactions applicable to a single budget.
func (h *AccountHandler) GetTransactionsByBudget(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetTransactionsByBudget] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	txns, err := h.container.TransactionSvc.GetByBudgetID(r.Context(), userID, budgetID)
	if err != nil {
		if errors.Is(err, transactions.ErrBudgetNotFound) {
			writeError(w, http.StatusNotFound, "budget not found")
			return
		}

		log.Printf("Error getting budget transactions: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch budget transactions")
		return
	}

	writeJSON(w, http.StatusOK, txns)
}

// UpdateBudget partially or fully updates a budget.
func (h *AccountHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	log.Printf("[UpdateBudget] %s %s", r.Method, r.URL.Path)

	budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	var req types.UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := sqlcdb.UpdateBudgetParams{
		ID: budgetID,
	}

	// Column2 = name (empty string means keep existing via COALESCE)
	if req.Name != nil {
		params.Column2 = *req.Name
	} else {
		params.Column2 = ""
	}

	// Column3 = whether to update category, Category = new value
	if req.Category != nil {
		params.Column3 = true
		params.Category = pgtype.Text{String: *req.Category, Valid: *req.Category != ""}
	} else if req.ClearCategory {
		params.Column3 = true
		params.Category = pgtype.Text{}
	}

	// Column5 = limit_amount
	if req.LimitAmount != nil {
		combined := strings.ReplaceAll(*req.LimitAmount, ".", "")
		limitAmount, err := strconv.Atoi(combined)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit_amount must be a valid number")
			return
		}
		params.Column5 = pgtype.Numeric{Valid: true, Int: big.NewInt(int64(limitAmount))}
	}

	// Column6 = budget_period (empty string means keep existing)
	if req.Period != nil {
		params.Column6 = *req.Period
	} else {
		params.Column6 = ""
	}

	// Column7 = start_date
	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
			return
		}
		params.Column7 = pgtype.Date{Time: startDate, Valid: true}
	}

	// Column8 = whether to update end_date, EndDate = new value
	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
			return
		}
		params.Column8 = true
		params.EndDate = pgtype.Date{Time: endDate, Valid: true}
	} else if req.ClearEndDate {
		params.Column8 = true
		params.EndDate = pgtype.Date{}
	}

	budget, err := h.container.BudgetSvc.UpdateBudget(r.Context(), params)
	if err != nil {
		log.Printf("Error updating budget: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to update budget")
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

// DeleteBudget removes a budget by ID.
func (h *AccountHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DeleteBudget] %s %s", r.Method, r.URL.Path)

	budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	err = h.container.BudgetSvc.DeleteBudget(r.Context(), budgetID)
	if err != nil {
		log.Printf("Error deleting budget: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to delete budget")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
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
