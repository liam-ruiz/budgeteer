package bank_accounts

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
	"github.com/liam-ruiz/budgeteer/internal/plaid"
	"github.com/liam-ruiz/budgeteer/internal/plaid_items"
	"github.com/liam-ruiz/budgeteer/internal/transactions"
	"github.com/liam-ruiz/budgeteer/internal/util"
	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

type TransactionUpdate = transactions.TransactionUpdate

var ErrAccountNotFound = errors.New("account not found")

// Service handles bank account business logic.
type Service struct {
	repo            Repository
	plaidItemSvc    *plaid_items.Service
	plaidApiSvc     *plaid.Service
	transactionsSvc *transactions.Service
}

// NewService creates a new bank account service.
func NewService(repo Repository, plaidItemSvc *plaid_items.Service, plaidApiSvc *plaid.Service, transactionsSvc *transactions.Service) *Service {
	return &Service{
		repo:            repo,
		plaidItemSvc:    plaidItemSvc,
		plaidApiSvc:     plaidApiSvc,
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

// GetAccount returns a single bank account owned by the user.
func (s *Service) GetAccount(ctx context.Context, plaidAccountID string) (AccountResponseWithUserID, error) {
	account, err := s.repo.GetByPlaidAccountID(ctx, plaidAccountID)
	if err != nil {
		return AccountResponseWithUserID{}, err
	}

	return ToAccountResponseWithUserID(account), nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID, plaidAccountID string) error {
	accounts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if account.PlaidAccountID == plaidAccountID {
			return s.repo.Delete(ctx, plaidAccountID)
		}
	}

	return ErrAccountNotFound
}

func ToAccountResponse(a sqlcdb.BankAccount) AccountResponse {
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

func ToAccountResponseWithUserID(a sqlcdb.GetBankAccountByPlaidAccountIDRow) AccountResponseWithUserID {
	subtype := ""
	if a.AccountSubtype.Valid {
		subtype = a.AccountSubtype.String
	}

	return AccountResponseWithUserID{
		PlaidItemID:      a.PlaidItemID,
		PlaidAccountID:   a.PlaidAccountID,
		AccountName:      a.AccountName,
		AccountType:      a.AccountType,
		AccountSubtype:   subtype,
		CurrentBalance:   util.NumericToString(a.CurrentBalance),
		AvailableBalance: util.NumericToString(a.AvailableBalance),
		IsoCurrencyCode:  a.IsoCurrencyCode,
		UserID:           a.AppUserID.String(),
	}
}

// CreatePlaidItem persists a new Plaid item (connection).
func (s *Service) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return s.repo.CreatePlaidItem(ctx, params)
}

// CreateBankAccount creates a new bank account under a Plaid item.
func (s *Service) CreateBankAccount(ctx context.Context, params sqlcdb.CreateBankAccountParams) (sqlcdb.BankAccount, error) {
	return s.repo.CreateBankAccount(ctx, params)
}

func (s *Service) LinkNewItem(ctx context.Context, appUserID uuid.UUID, plaidItemID, token string, resp *plaidlib.AccountsGetResponse) error {
	// 1. Create the Plaid Item
	item := resp.GetItem()
	plaidItem, err := s.plaidItemSvc.CreatePlaidItem(ctx, sqlcdb.CreatePlaidItemParams{
		AppUserID:        appUserID,
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
			PlaidItemID:      plaidItem.PlaidItemID,
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

func (s *Service) SyncLinkedAccounts(ctx context.Context, userID uuid.UUID) error {
	items, err := s.repo.GetPlaidItemsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(items))
	for _, item := range items {
		item := item
		wg.Go(func() {

			if err := s.refreshAccounts(ctx, item); err != nil {
				errCh <- fmt.Errorf("refresh accounts for item %s: %w", item.PlaidItemID, err)
				return
			}

			cursor := ""
			if item.PlaidCursor.Valid {
				cursor = item.PlaidCursor.String
			}
			if err := s.SyncTransactions(ctx, item.PlaidItemID, cursor); err != nil {
				errCh <- fmt.Errorf("sync transactions for item %s: %w", item.PlaidItemID, err)
			}
		})
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

func (s *Service) refreshAccounts(ctx context.Context, item sqlcdb.PlaidItem) error {
	accountsReq := plaidlib.NewAccountsGetRequest(item.PlaidAccessToken)
	accountsResp, err := s.plaidApiSvc.GetAccounts(ctx, accountsReq)
	if err != nil {
		return err
	}

	for _, acc := range accountsResp.GetAccounts() {
		balance := acc.GetBalances()
		_, err := s.repo.UpsertBankAccount(ctx, sqlcdb.UpsertBankAccountParams{
			PlaidItemID:      item.PlaidItemID,
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
	if err != nil {
		log.Printf("[SyncTransactions] failed to get plaid item %s: %v", plaidItemID, err)
		return err
	}

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
		if err != nil {
			log.Printf("[SyncTransactions] plaid sync API error for item %s: %v", plaidItemID, err)
			return err
		}

		added = append(added, resp.GetAdded()...)
		modified = append(modified, resp.GetModified()...)
		removed = append(removed, resp.GetRemoved()...)

		hasMore = resp.GetHasMore()
		cursor = resp.GetNextCursor()
	}

	log.Printf("[SyncTransactions] item %s: %d added, %d modified, %d removed",
		plaidItemID, len(added), len(modified), len(removed))

	transactionsUpdate := TransactionUpdate{
		Added:       added,
		Modified:    modified,
		Removed:     removed,
		Cursor:      cursor,
		PlaidItemID: acc.PlaidItemID,
	}

	// persist everything
	err = s.transactionsSvc.SyncTransactions(ctx, transactionsUpdate)
	if err != nil {
		log.Printf("[SyncTransactions] failed to persist transactions for item %s: %v", plaidItemID, err)
		return err
	}

	// update cursor
	err = s.repo.UpdateCursor(ctx, plaidItemID, cursor)
	if err != nil {
		log.Printf("[SyncTransactions] failed to update cursor for item %s: %v", plaidItemID, err)
		return err
	}

	return nil
}
