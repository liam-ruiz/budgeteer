package plaid_items

import (
	"context"

	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
)

type Repository interface {
	GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error)
	CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error)
}

type repository struct {
	q *sqlcdb.Queries
}

func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error) {
	return r.q.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
}

func (r *repository) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return r.q.CreatePlaidItem(ctx, params)
}