package transactions

import plaidlib "github.com/plaid/plaid-go/v20/plaid"

type TransactionUpdate struct {
	PlaidItemID string
	Cursor string
	Added []plaidlib.Transaction
	Modified []plaidlib.Transaction
	Removed []plaidlib.RemovedTransaction
}