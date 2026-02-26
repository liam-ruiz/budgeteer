package bank_accounts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
)

// Repository defines the interface for bank account data access.
type Repository interface {
	CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error)
	CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (BankAccount, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]BankAccount, error)
	GetByPlaidAccountID(ctx context.Context, plaidAccountID string) (BankAccount, error)
	GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error)
	UpdateBalance(ctx context.Context, params sqlcdb.UpdateBankAccountBalanceParams) error
	UpsertBankAccount(ctx context.Context, params sqlcdb.UpsertBankAccountParams) (BankAccount, error)
	UpdateCursor(ctx context.Context, plaidItemID string, cursor string) error
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

func (r *repository) UpsertBankAccount(ctx context.Context, params sqlcdb.UpsertBankAccountParams) (BankAccount, error) {
	row, err := r.q.UpsertBankAccount(ctx, params)
	if err != nil {
		return BankAccount{}, err
	}
	return toBankAccount(row), nil
}

func (r *repository) CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (BankAccount, error) {
	row, err := r.q.CreateBankAccount(ctx, params)
	if err != nil {
		return BankAccount{}, err
	}
	return toBankAccount(row), nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]BankAccount, error) {
	rows, err := r.q.GetBankAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	accounts := make([]BankAccount, len(rows))
	for i, row := range rows {
		accounts[i] = toBankAccount(row)
	}
	return accounts, nil
}


func (r *repository) GetByPlaidAccountID(ctx context.Context, plaidAccountID string) (BankAccount, error) {
	row, err := r.q.GetBankAccountByPlaidAccountID(ctx, plaidAccountID)
	if err != nil {
		return BankAccount{}, err
	}
	return toBankAccount(row), nil
}

func (r *repository) GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error) {
	return r.q.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
}

func (r *repository) UpdateBalance(ctx context.Context, params sqlcdb.UpdateBankAccountBalanceParams) error {
	return r.q.UpdateBankAccountBalance(ctx, params)
}

func toBankAccount(row sqlcdb.BankAccount) BankAccount {
	return BankAccount{
		PlaidItemID:      row.PlaidItemID,
		PlaidAccountID:   row.PlaidAccountID,
		AccountName:      row.AccountName,
		OfficialName:     row.OfficialName,
		AccountType:      row.AccountType,
		AccountSubtype:   row.AccountSubtype,
		CurrentBalance:   row.CurrentBalance,
		AvailableBalance: row.AvailableBalance,
		IsoCurrencyCode:  row.IsoCurrencyCode,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func (r *repository) UpdateCursor(ctx context.Context, plaidItemID string, cursor string) error {
	valid := cursor != ""
	return r.q.UpdatePlaidItemCursor(ctx, sqlcdb.UpdatePlaidItemCursorParams{
		PlaidItemID: plaidItemID,
		PlaidCursor: pgtype.Text{String: cursor, Valid: valid},
	})
}

