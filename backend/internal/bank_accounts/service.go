package bank_accounts

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budget/internal/bank_accounts/plaid_items"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/plaid"

	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

// Service handles bank account business logic.
type Service struct {
	repo Repository
	plaidItemService *plaid_items.Service
	plaidApiService *plaid.Service
}

// NewService creates a new bank account service.
func NewService(repo Repository, plaidItemService *plaid_items.Service, plaidApiService *plaid.Service) *Service {
	return &Service{
		repo: repo,
		plaidItemService: plaidItemService,
		plaidApiService: plaidApiService,
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

// CreatePlaidItem persists a new Plaid item (connection).
func (s *Service) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return s.repo.CreatePlaidItem(ctx, params)
}

// CreateBankAccount creates a new bank account under a Plaid item.
func (s *Service) CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (BankAccount, error) {
	return s.repo.CreateBankAccount(ctx, params)
}

func (s *Service) LinkNewItem(ctx context.Context, userID uuid.UUID, plaidItemID, token string, resp *plaidlib.AccountsGetResponse) error {
    // 1. Create the Plaid Item
	item := resp.GetItem()
    plaidItem, err := s.plaidItemService.CreatePlaidItem(ctx, sqlcdb.CreatePlaidItemParams{
        UserID:           userID,
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
            ItemID:           plaidItem.ID,
            PlaidAccountID:   acc.GetAccountId(),
            AccountName:      acc.GetName(),
            OfficialName:     sql.NullString{String: acc.GetOfficialName(), Valid: acc.OfficialName.IsSet()},
            AccountType:      string(acc.GetType()),
            AccountSubtype:   sql.NullString{String: string(acc.GetSubtype()), Valid: acc.Subtype.IsSet()},
            CurrentBalance:   fmt.Sprintf("%f", balance.GetCurrent()),
            AvailableBalance: fmt.Sprintf("%f", balance.GetAvailable()),
            IsoCurrencyCode:  balance.GetIsoCurrencyCode(),
        })
        if err != nil {
            return err
        }
    }
    return nil
}
// SyncTransactions syncs transactions for a specific bank account. Cursor 
// defines where in the transaction history to fetch from, `""` for from the start
func (s *Service) SyncTransactions(ctx context.Context, plaidItemID string, cursor string) (TransactionUpdate, error) {
    // get the access token from DB
    acc, err := s.repo.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
    if err != nil { return TransactionUpdate{}, err }

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

        resp, err := s.plaidApiService.SyncTransactions(ctx, request)
        if err != nil { return TransactionUpdate{}, err }

        added = append(added, resp.GetAdded()...)
        modified = append(modified, resp.GetModified()...)
        removed = append(removed, resp.GetRemoved()...)
        
        hasMore = resp.GetHasMore()
        cursor = resp.GetNextCursor()
    }

    // 3. Persist everything in a single Transaction
    return TransactionUpdate{
        PlaidItemID: plaidItemID,
        Cursor: cursor,
        Added: added,
        Modified: modified,
        Removed: removed,
    }, nil
}
