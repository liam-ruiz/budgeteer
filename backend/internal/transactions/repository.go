package transactions

import (
	"context"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/util"
)

// Repository defines the interface for transaction data access.
type Repository interface {
	Create(ctx context.Context, params sqlcdb.CreateTransactionParams) (Transaction, error)
	Upsert(ctx context.Context, params sqlcdb.UpsertTransactionParams) (Transaction, error)
	GetByAccountID(ctx context.Context, plaidAccountID string) ([]Transaction, error)
	GetByUserID(ctx context.Context, appUserID uuid.UUID) ([]Transaction, error)
}

type repository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new transaction repository backed by sqlc queries.
func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, params sqlcdb.CreateTransactionParams) (Transaction, error) {
	row, err := r.q.CreateTransaction(ctx, params)
	if err != nil {
		return Transaction{}, err
	}
	return toTransaction(row), nil
}

func (r *repository) Upsert(ctx context.Context, params sqlcdb.UpsertTransactionParams) (Transaction, error) {
	row, err := r.q.UpsertTransaction(ctx, params)
	if err != nil {
		return Transaction{}, err
	}
	return toTransaction(row), nil
}

func (r *repository) GetByAccountID(ctx context.Context, plaidAccountID string) ([]Transaction, error) {
	rows, err := r.q.GetTransactionsByAccountID(ctx, plaidAccountID)
	if err != nil {
		return nil, err
	}
	return toTransactions(rows), nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Transaction, error) {
	rows, err := r.q.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toTransactions(rows), nil
}

func toTransaction(row sqlcdb.Transaction) Transaction {
	return Transaction{
		PlaidTransactionID:      row.PlaidTransactionID,
		AccountID:               row.PlaidAccountID,
		Date:                    row.TransactionDate.Time.Format("2006-01-02"),
		Name:                    row.TransactionName,
		Amount:                  util.NumericToString(row.Amount),
		Pending:                 row.Pending,
		MerchantName:            row.MerchantName.String,
		LogoUrl:                 row.LogoUrl.String,
		PersonalFinanceCategory: row.PersonalFinanceCategory.String,
		DetailedCategory:        row.DetailedCategory.String,
		CategoryConfidenceLevel: row.CategoryConfidenceLevel.String,
		CategoryIconUrl:         row.CategoryIconUrl.String,
		CreatedAt:               row.CreatedAt.Time,
	}
}

func toTransactions(rows []sqlcdb.Transaction) []Transaction {
	out := make([]Transaction, len(rows))
	for i, row := range rows {
		out[i] = toTransaction(row)
	}
	return out
}
