package transactions

import (
	"time"


)

// Transaction represents a financial transaction from a linked account.
type Transaction struct {
	PlaidTransactionID string	`json:"transaction_id"`
	AccountID string 			`json:"account_id"`
	Date      string 			`json:"date"`
	Name      string    		`json:"name"`
	Category  string    		`json:"category"`
	Amount    string    		`json:"amount"`
	Pending   bool      		`json:"pending"`
	CreatedAt time.Time 		`json:"created_at"`
}
