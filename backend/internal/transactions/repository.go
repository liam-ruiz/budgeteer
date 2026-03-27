package transactions

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
)

// Repository defines the interface for transaction data access.
type Repository interface {
	Create(ctx context.Context, params sqlcdb.CreateTransactionParams) (sqlcdb.Transaction, error)
	Upsert(ctx context.Context, params sqlcdb.UpsertTransactionParams) (sqlcdb.Transaction, error)
	GetByAccountID(ctx context.Context, plaidAccountID string) ([]sqlcdb.GetTransactionsByAccountIDRow, error)
	GetByUserID(ctx context.Context, appUserID uuid.UUID) ([]sqlcdb.GetTransactionsByUserIDRow, error)
	GetByBudgetID(ctx context.Context, budgetID uuid.UUID) ([]sqlcdb.GetTransactionsByBudgetIDRow, error)
	UpdateCategory(ctx context.Context, params sqlcdb.UpdateTransactionCategoryParams) error
	Delete(ctx context.Context, plaidTransactionID string) error
}

type repository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new transaction repository backed by sqlc queries.
func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, params sqlcdb.CreateTransactionParams) (sqlcdb.Transaction, error) {
	return r.q.CreateTransaction(ctx, params)
}

func (r *repository) Upsert(ctx context.Context, params sqlcdb.UpsertTransactionParams) (sqlcdb.Transaction, error) {
	log.Printf("Upserting transaction: %+v", params)
	return r.q.UpsertTransaction(ctx, params)
}

func (r *repository) GetByAccountID(ctx context.Context, plaidAccountID string) ([]sqlcdb.GetTransactionsByAccountIDRow, error) {
	return r.q.GetTransactionsByAccountID(ctx, plaidAccountID)
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]sqlcdb.GetTransactionsByUserIDRow, error) {
	return r.q.GetTransactionsByUserID(ctx, userID)

}

func (r *repository) GetByBudgetID(ctx context.Context, budgetID uuid.UUID) ([]sqlcdb.GetTransactionsByBudgetIDRow, error) {
	return r.q.GetTransactionsByBudgetID(ctx, budgetID)
}

func (r *repository) UpdateCategory(ctx context.Context, params sqlcdb.UpdateTransactionCategoryParams) error {
	return r.q.UpdateTransactionCategory(ctx, params)
}

func (r *repository) Delete(ctx context.Context, plaidTransactionID string) error {
	return r.q.DeleteTransaction(ctx, plaidTransactionID)
}
