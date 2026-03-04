package transactions

import (
	"time"
)

// Transaction represents a financial transaction from a linked account.
type Transaction struct {
	PlaidTransactionID      string    `json:"transaction_id"`
	AccountID               string    `json:"account_id"`
	Date                    string    `json:"date"`
	Name                    string    `json:"name"`
	Amount                  string    `json:"amount"`
	Pending                 bool      `json:"pending"`
	MerchantName            string    `json:"merchant_name,omitempty"`
	LogoUrl                 string    `json:"logo_url,omitempty"`
	PersonalFinanceCategory string    `json:"personal_finance_category,omitempty"`
	DetailedCategory        string    `json:"detailed_category,omitempty"`
	CategoryConfidenceLevel string    `json:"category_confidence_level,omitempty"`
	CategoryIconUrl         string    `json:"category_icon_url,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
}

type TransactionWithAccountName struct {
	PlaidTransactionID      string    `json:"transaction_id"`
	AccountID               string    `json:"account_id"`
	Date                    string    `json:"date"`
	Name                    string    `json:"name"`
	Amount                  string    `json:"amount"`
	Pending                 bool      `json:"pending"`
	MerchantName            string    `json:"merchant_name,omitempty"`
	LogoUrl                 string    `json:"logo_url,omitempty"`
	PersonalFinanceCategory string    `json:"personal_finance_category,omitempty"`
	DetailedCategory        string    `json:"detailed_category,omitempty"`
	CategoryConfidenceLevel string    `json:"category_confidence_level,omitempty"`
	CategoryIconUrl         string    `json:"category_icon_url,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	AccountName string `json:"account_name"`
}

type TransactionWithAccountNameResponse struct {
	TransactionWithAccountName
	UserID string `json:"user_id"`
}
