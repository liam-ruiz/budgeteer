package bank_accounts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budget/internal/bank_accounts/plaid_items"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/plaid"
	"github.com/liam-ruiz/budget/internal/transactions"
	"github.com/liam-ruiz/budget/internal/util"

	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

type TransactionUpdate = transactions.TransactionUpdate

// Service handles bank account business logic.
type Service struct {
	repo Repository
	plaidItemSvc *plaid_items.Service
	plaidApiSvc *plaid.Service
	transactionsSvc *transactions.Service
}

// NewService creates a new bank account service.
func NewService(repo Repository, plaidItemSvc *plaid_items.Service, plaidApiSvc *plaid.Service, transactionsSvc *transactions.Service) *Service {
	return &Service{
		repo: repo,
		plaidItemSvc: plaidItemSvc,
		plaidApiSvc: plaidApiSvc,
		transactionsSvc: transactionsSvc,
	}
}

// GetAccounts returns all bank accounts for a user.
func (s *Service) GetAccounts(ctx context.Context, userID uuid.UUID) ([]AccountResponse, error) {
	accounts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]AccountResponse, len(accounts))
	for i, a := range accounts {
		out[i] = ToAccountResponse(a)
	}
	return out, nil
}

func ToAccountResponse(a BankAccount) AccountResponse {
	subtype := ""
	if a.AccountSubtype.Valid {
		subtype = a.AccountSubtype.String
	}

	return AccountResponse{
		PlaidItemID:      a.PlaidItemID,
		PlaidAccountID:   a.PlaidAccountID,
		AccountName:      a.AccountName,
		AccountType:      a.AccountType,
		AccountSubtype:   subtype,
		CurrentBalance:   util.NumericToString(a.CurrentBalance),
		AvailableBalance: util.NumericToString(a.AvailableBalance),
		IsoCurrencyCode:  a.IsoCurrencyCode,
	}
}

// CreatePlaidItem persists a new Plaid item (connection).
func (s *Service) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return s.repo.CreatePlaidItem(ctx, params)
}

// CreateBankAccount creates a new bank account under a Plaid item.
func (s *Service) CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (BankAccount, error) {
	return s.repo.CreateBankAccount(ctx, params)
}

func (s *Service) LinkNewItem(ctx context.Context, appUserID uuid.UUID, plaidItemID, token string, resp *plaidlib.AccountsGetResponse) error {
    // 1. Create the Plaid Item
	item := resp.GetItem()
    plaidItem, err := s.plaidItemSvc.CreatePlaidItem(ctx, sqlcdb.CreatePlaidItemParams{
        AppUserID:           appUserID,
        PlaidItemID:      plaidItemID,
        PlaidAccessToken: token,
        InstitutionName:  item.GetInstitutionId(), // Or pass name from frontend
    })
    if err != nil {
        return err
    }

    // 2. Loop and Upsert each account found in the Link session
    for _, acc := range resp.GetAccounts() {
		balance := acc.GetBalances()
        _, err := s.repo.UpsertBankAccount(ctx, sqlcdb.UpsertBankAccountParams{
            PlaidItemID:           plaidItem.PlaidItemID,
            PlaidAccountID:   acc.GetAccountId(),
            AccountName:      acc.GetName(),
            OfficialName:     pgtype.Text{String: acc.GetOfficialName(), Valid: acc.OfficialName.IsSet()},
            AccountType:      string(acc.GetType()),
            AccountSubtype:   pgtype.Text{String: string(acc.GetSubtype()), Valid: acc.Subtype.IsSet()},
            CurrentBalance:   util.Float64ToNumeric(balance.GetCurrent()),
            AvailableBalance: util.Float64ToNumeric(balance.GetAvailable()),
            IsoCurrencyCode:  acc.Balances.GetIsoCurrencyCode(),
        })
        if err != nil {
            return err
        }
    }
    return nil
}
// SyncTransactions syncs transactions for a specific bank account. Cursor 
// defines where in the transaction history to fetch from, `""` for from the start
func (s *Service) SyncTransactions(ctx context.Context, plaidItemID string, cursor string) error {
    // get the access token from DB
    acc, err := s.repo.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
    if err != nil { return err }

    var (
        added    = make([]plaidlib.Transaction, 0)
        modified = make([]plaidlib.Transaction, 0)
        removed  = make([]plaidlib.RemovedTransaction, 0)
        hasMore  = true
    )

    // 2. Loop until we have all updates
    for hasMore {
        request := plaidlib.NewTransactionsSyncRequest(acc.PlaidAccessToken)
        if cursor != "" {
            request.SetCursor(cursor)
        }

        resp, err := s.plaidApiSvc.SyncTransactions(ctx, request)
        if err != nil { return err }

        added = append(added, resp.GetAdded()...)
        modified = append(modified, resp.GetModified()...)
        removed = append(removed, resp.GetRemoved()...)
        
        hasMore = resp.GetHasMore()
        cursor = resp.GetNextCursor()
    }
	transactionsUpdate := TransactionUpdate{
		Added:    added,
		Modified: modified,
		Removed:  removed,
		Cursor:   cursor,
		PlaidItemID: acc.PlaidItemID,
	}

    // persist everything in a single Transaction
	err = s.transactionsSvc.CreateTransactions(ctx, transactionsUpdate)
    if err != nil { return err }
    
	// update cursor
	err = s.repo.UpdateCursor(ctx, plaidItemID, cursor)
    if err != nil { return err }

	return nil
}
