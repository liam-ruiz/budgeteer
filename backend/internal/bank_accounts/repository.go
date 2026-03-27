package bank_accounts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
)

// Repository defines the interface for bank account data access.
type Repository interface {
	CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error)
	CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (sqlcdb.BankAccount, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]sqlcdb.BankAccount, error)
	GetByPlaidAccountID(ctx context.Context, plaidAccountID string) (sqlcdb.GetBankAccountByPlaidAccountIDRow, error)
	GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error)
	UpdateBalance(ctx context.Context, params sqlcdb.UpdateBankAccountBalanceParams) error
	UpsertBankAccount(ctx context.Context, params sqlcdb.UpsertBankAccountParams) (sqlcdb.BankAccount, error)
	UpdateCursor(ctx context.Context, plaidItemID string, cursor string) error
	Delete(ctx context.Context, plaidAccountID string) error
}

type repository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new bank account repository backed by sqlc queries.
func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return r.q.CreatePlaidItem(ctx, params)
}

func (r *repository) UpsertBankAccount(ctx context.Context, params sqlcdb.UpsertBankAccountParams) (sqlcdb.BankAccount, error) {
	return r.q.UpsertBankAccount(ctx, params)
}

func (r *repository) CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (sqlcdb.BankAccount, error) {
	return r.q.CreateBankAccount(ctx, params)
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]sqlcdb.BankAccount, error) {
	return r.q.GetBankAccountsByUserID(ctx, userID)
}

func (r *repository) GetByPlaidAccountID(ctx context.Context, plaidAccountID string) (sqlcdb.GetBankAccountByPlaidAccountIDRow, error) {
	return r.q.GetBankAccountByPlaidAccountID(ctx, plaidAccountID)
}

func (r *repository) GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error) {
	return r.q.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
}

func (r *repository) UpdateBalance(ctx context.Context, params sqlcdb.UpdateBankAccountBalanceParams) error {
	return r.q.UpdateBankAccountBalance(ctx, params)
}

func (r *repository) UpdateCursor(ctx context.Context, plaidItemID string, cursor string) error {
	valid := cursor != ""
	return r.q.UpdatePlaidItemCursor(ctx, sqlcdb.UpdatePlaidItemCursorParams{
		PlaidItemID: plaidItemID,
		PlaidCursor: pgtype.Text{String: cursor, Valid: valid},
	})
}

func (r *repository) Delete(ctx context.Context, plaidAccountID string) error {
	return r.q.DeleteBankAccount(ctx, plaidAccountID)
}
