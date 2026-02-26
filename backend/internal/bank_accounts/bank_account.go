package bank_accounts

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// BankAccount represents a bank account linked via Plaid.
type BankAccount struct {
	PlaidItemID      string      `json:"item_id"`
	PlaidAccountID   string         `json:"-"`
	AccountName      string         `json:"account_name"`
	OfficialName     pgtype.Text `json:"official_name"`
	AccountType      string         `json:"account_type"`
	AccountSubtype   pgtype.Text `json:"account_subtype"`
	CurrentBalance   pgtype.Numeric         `json:"current_balance"`
	AvailableBalance pgtype.Numeric         `json:"available_balance"`
	IsoCurrencyCode  string         `json:"iso_currency_code"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// AccountResponse is the JSON-safe view sent to clients.
type AccountResponse struct {
	PlaidItemID      string 	`json:"item_id"`
	PlaidAccountID   string 	`json:"account_id"`
	AccountName      string    `json:"account_name"`
	AccountType      string    `json:"account_type"`
	AccountSubtype   string    `json:"account_subtype,omitempty"`
	CurrentBalance   string    `json:"current_balance"`
	AvailableBalance string    `json:"available_balance"`
	IsoCurrencyCode  string    `json:"iso_currency_code"`
}


