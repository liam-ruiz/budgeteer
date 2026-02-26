package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

// Handler holds all service dependencies for the API.
type Handler struct {
	container *dependencies.Container
}

// NewHandler creates a new API handler with all service dependencies.
func NewHandler(
	container *dependencies.Container,
) *Handler {
	return &Handler{
		container: container,
	}
}

// Register creates a new user account.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Default().Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.container.UserSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Default().Printf("Error registering user: %v\n", err)
		writeError(w, http.StatusConflict, "email already in use")
		return
	}

	token, err := auth.GenerateJWT(user.ID, h.container.Cfg.JWTSecret)
	if err != nil {
		log.Default().Printf("Error generating JWT: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, types.AuthResponse{
		Token: token,
		User: types.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

// Login authenticates a user and returns a JWT.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Default().Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.container.UserSvc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Default().Printf("Error authenticating user: %v\n", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateJWT(user.ID, h.container.Cfg.JWTSecret)
	if err != nil {
		log.Default().Printf("Error generating JWT: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, types.AuthResponse{
		Token: token,
		User: types.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

// GetAccounts returns all linked bank accounts for the authenticated user.
func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Default().Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.container.AccountSvc.GetAccounts(r.Context(), userID)
	if err != nil {
		log.Default().Printf("Error getting accounts: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch accounts")
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

// GetTransactions returns all transactions for the authenticated user.
func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Default().Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	txns, err := h.container.TransactionSvc.GetByUser(r.Context(), userID)
	if err != nil {
		log.Default().Printf("Error getting transactions: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	writeJSON(w, http.StatusOK, txns)
}

// CreateBudget creates a new budget for the authenticated user.
func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Default().Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req types.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Default().Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Category == "" || req.LimitAmount == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "category, limit_amount, and start_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		log.Default().Printf("Error parsing start date: %v\n", err)
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
		log.Default().Printf("Error parsing limit amount: %v\n", err)
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
			log.Default().Printf("Error parsing end date: %v\n", err)
			writeError(w, http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
			return
		}
		params.EndDate = pgtype.Date{Time: endDate, Valid: true}
	}

	budget, err := h.container.BudgetSvc.CreateBudget(r.Context(), params)
	if err != nil {
		log.Default().Printf("Error creating budget: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to create budget")
		return
	}

	writeJSON(w, http.StatusCreated, budget)
}

// GetBudgets returns all budgets for the authenticated user.
func (h *Handler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Default().Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.container.BudgetSvc.GetBudgets(r.Context(), userID)
	if err != nil {
		log.Default().Printf("Error getting budgets: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch budgets")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// --- Plaid Handlers ---

func (h *Handler) ExchangePlaidPublicToken(w http.ResponseWriter, r *http.Request) {
    userID, _ := auth.GetUserID(r.Context()) // Assuming middleware handled this

    var req types.ExchangeTokenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // exchange public token for access token
    exchangeReq := plaidlib.NewItemPublicTokenExchangeRequest(req.PublicToken)
    resp, err := h.container.PlaidAPISvc.ExchangePublicToken(r.Context(), exchangeReq)
    if err != nil {
        writeError(w, http.StatusBadGateway, "plaid exchange failed")
        return
    }

    accessToken := resp.GetAccessToken()
    plaidItemID := resp.GetItemId()

    // fetch metadata for this Item
    // best to get the account details immediately so the user sees them.
    accountsReq := plaidlib.NewAccountsGetRequest(accessToken)
    accountsResp, err := h.container.PlaidAPISvc.GetAccountInfo(r.Context(), accountsReq)
    if err != nil {
        writeError(w, http.StatusBadGateway, "failed to fetch account details")
        return
    }

    // persist the Item and Accounts
    // We do this in the service layer to handle the transaction/upsert logic
    err = h.container.AccountSvc.LinkNewItem(r.Context(), userID, plaidItemID, accessToken, &accountsResp)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to save linked accounts")
        return
    }

    // trigger background transaction sync (Production-lite approach)
    // empty string for cursor means we want to sync all available transactions
    go h.container.AccountSvc.SyncTransactions(context.Background(), plaidItemID, "")

    writeJSON(w, http.StatusCreated, map[string]string{"status": "syncing"})
}

// CreateLinkToken generates a Plaid Link token for the authenticated user.
func (h *Handler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Default().Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	products := []plaidlib.Products{plaidlib.PRODUCTS_TRANSACTIONS}
	countryCodes := []plaidlib.CountryCode{plaidlib.COUNTRYCODE_US}

	user := plaidlib.LinkTokenCreateRequestUser{
		ClientUserId: userID.String(),
	}

	linkReq := plaidlib.NewLinkTokenCreateRequest(
		"Budget",
		"en",
		countryCodes,
		user,
	)
	linkReq.SetProducts(products)

	resp, err := h.container.PlaidAPISvc.CreateLinkToken(r.Context(), linkReq)
	if err != nil {
		log.Default().Printf("Error creating link token: %v\n", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create link token: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, types.CreateLinkTokenResponse{
		LinkToken: resp.GetLinkToken(),
	})
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
