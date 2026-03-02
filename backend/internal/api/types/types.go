package types

import "github.com/google/uuid"

// AuthRequest is used for both login and registration.
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after a successful login or registration.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// CreateBudgetRequest is the payload for creating a new budget.
type CreateBudgetRequest struct {
	Category    string  `json:"category"`
	LimitAmount string  `json:"limit_amount"`
	Period      string  `json:"period"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// --- Plaid ---

// ExchangeTokenRequest is the payload for exchanging a Plaid public token.
type ExchangeTokenRequest struct {
	PublicToken     string `json:"public_token"`
	InstitutionName string `json:"institution_name"`
	AccountName     string `json:"account_name"`
	AccountType     string `json:"account_type"`
}

// CreateLinkTokenResponse is returned when a Plaid link token is created.
type CreateLinkTokenResponse struct {
	LinkToken string `json:"link_token"`
}

// ExchangeTokenResponse is returned after a successful Plaid token exchange.
type ExchangeTokenResponse struct {
	AccountID uuid.UUID `json:"account_id"`
	ItemID    string    `json:"item_id"`
}

// ErrorResponse is returned on any handler error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// PlaidWebhook is the payload for a Plaid webhook.
type PlaidWebhook struct {
	WebhookType string `json:"webhook_type"`
	WebhookCode string `json:"webhook_code"`
	PlaidItemID string `json:"item_id"`
}