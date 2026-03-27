package transactions

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
	"github.com/liam-ruiz/budgeteer/internal/util"
	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

var ErrTransactionNotFound = errors.New("transaction not found")
var ErrAccountNotFound = errors.New("account not found")
var ErrBudgetNotFound = errors.New("budget not found")
var ErrInvalidTransactionCategory = errors.New("invalid transaction category")

var allowedTransactionCategories = map[string]struct{}{
	"INCOME":                    {},
	"LOAN_DISBURSEMENTS":        {},
	"LOAN_PAYMENTS":             {},
	"TRANSFER_IN":               {},
	"TRANSFER_OUT":              {},
	"BANK_FEES":                 {},
	"ENTERTAINMENT":             {},
	"FOOD_AND_DRINK":            {},
	"GENERAL_MERCHANDISE":       {},
	"HOME_IMPROVEMENT":          {},
	"MEDICAL":                   {},
	"PERSONAL_CARE":             {},
	"GENERAL_SERVICES":          {},
	"GOVERNMENT_AND_NON_PROFIT": {},
	"TRANSPORTATION":            {},
	"TRAVEL":                    {},
	"RENT_AND_UTILITIES":        {},
	"PERSONAL":                  {},
	"OTHER":                     {},
}

// Service handles transaction business logic.
type Service struct {
	repo Repository
}

// NewService creates a new transaction service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetByUser returns all transactions across all linked accounts for a user.
func (s *Service) GetByUser(ctx context.Context, userID uuid.UUID) ([]TransactionWithAccountName, error) {
	DBTransactions, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toTransactionsWithAccountName(DBTransactions), nil
}

// GetByAccount returns all transactions for a single account.
func (s *Service) GetByAccount(ctx context.Context, accountID string) ([]Transaction, error) {
	DBTransactions, err := s.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return toTransactions(DBTransactions), nil
}

// GetByBudget returns all user transactions applicable to the given budget.
func (s *Service) GetByBudgetID(ctx context.Context, userID, budgetID uuid.UUID) ([]TransactionWithAccountName, error) {
	DBTransactions, err := s.repo.GetByBudgetID(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	return toTransactionsWithAccountNameByBudgetID(DBTransactions), nil
}

// DeleteTransaction deletes a transaction with the given Plaid transaction ID if it belongs to the user.
func (s *Service) DeleteTransaction(ctx context.Context, userID uuid.UUID, plaidTransactionID string) error {
	transactions, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// TODO: optimize this by adding a GetTransactionByPlaidTransactionID query and verifying the transaction belongs to the user in SQL instead of in application code
	for _, transaction := range transactions {
		if transaction.PlaidTransactionID == plaidTransactionID {
			return s.repo.Delete(ctx, plaidTransactionID)
		}
	}

	return ErrTransactionNotFound
}

func (s *Service) UpdateCategory(ctx context.Context, userID uuid.UUID, plaidTransactionID, category string) error {
	category = strings.ToUpper(strings.TrimSpace(category))
	if _, ok := allowedTransactionCategories[category]; !ok {
		return ErrInvalidTransactionCategory
	}

	transactions, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		if transaction.PlaidTransactionID == plaidTransactionID {
			return s.repo.UpdateCategory(ctx, sqlcdb.UpdateTransactionCategoryParams{
				PlaidTransactionID: plaidTransactionID,
				UserPersonalFinanceCategory: pgtype.Text{
					String: category,
					Valid:  true,
				},
			})
		}
	}

	return ErrTransactionNotFound
}

// CreateTransaction persists a single transaction.
func (s *Service) CreateTransaction(ctx context.Context, params sqlcdb.CreateTransactionParams) (Transaction, error) {
	dbTxn, err := s.repo.Create(ctx, params)
	if err != nil {
		return Transaction{}, err
	}
	return toCreatedTransaction(dbTxn), nil
}

// CreateTransactions upserts transactions from a Plaid sync update.
// Both Added and Modified transactions are upserted into the database.
func (s *Service) SyncTransactions(ctx context.Context, update TransactionUpdate) error {
	allTransactions := make([]plaidlib.Transaction, 0, len(update.Added)+len(update.Modified))
	allTransactions = append(allTransactions, update.Added...)
	allTransactions = append(allTransactions, update.Modified...)

	if len(allTransactions) == 0 {
		log.Printf("[SyncTransactions] no transactions to upsert for item %s", update.PlaidItemID)
		return nil
	}
	// TODO: optimize this by doing batch upserts in the repository instead of upserting transactions one by one in application code
	// upsert added and modified
	err := s.upsertTransactions(ctx, allTransactions)

	if err != nil {
		return err
	}

	// delete removed
	deleteErrors := 0
	for _, t := range update.Removed {
		id, ok := t.GetTransactionIdOk()
		if !ok {
			log.Printf("[SyncTransactions] missing transaction ID for removed transaction in item %s", update.PlaidItemID)
			continue
		}
		err := s.repo.Delete(ctx, *id)
		if err != nil {
			deleteErrors++
			log.Printf("[SyncTransactions] failed to delete transaction %s: %v", *id, err)
			// continue deleting the rest of the removed transactions even if one delete fails
		}
	}
	if deleteErrors > 0 {
		log.Printf("[SyncTransactions] failed to delete %d/%d removed transactions for item %s", deleteErrors, len(update.Removed), update.PlaidItemID)
		return fmt.Errorf("failed to delete %d/%d removed transactions for item %s", deleteErrors, len(update.Removed), update.PlaidItemID)
	}

	log.Printf("[SyncTransactions] successfully upserted %d transactions for item %s", len(allTransactions), update.PlaidItemID)
	return nil
}

func toTransactionWithAccountName(row sqlcdb.GetTransactionsByUserIDRow) TransactionWithAccountName {
	return TransactionWithAccountName{
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
		AccountName:             row.AccountName,
	}
}

func toTransactionWithAccountNameByBudgetID(row sqlcdb.GetTransactionsByBudgetIDRow) TransactionWithAccountName {
	return TransactionWithAccountName{
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
		AccountName:             row.AccountName,
	}
}

func toTransactionsWithAccountName(rows []sqlcdb.GetTransactionsByUserIDRow) []TransactionWithAccountName {
	out := make([]TransactionWithAccountName, len(rows))
	for i, row := range rows {
		out[i] = toTransactionWithAccountName(row)
	}
	return out
}

func toTransactionsWithAccountNameByBudgetID(rows []sqlcdb.GetTransactionsByBudgetIDRow) []TransactionWithAccountName {
	out := make([]TransactionWithAccountName, len(rows))
	for i, row := range rows {
		out[i] = toTransactionWithAccountNameByBudgetID(row)
	}
	return out
}

func toCreatedTransaction(row sqlcdb.Transaction) Transaction {
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

func toTransaction(row sqlcdb.GetTransactionsByAccountIDRow) Transaction {
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

func toTransactions(rows []sqlcdb.GetTransactionsByAccountIDRow) []Transaction {
	out := make([]Transaction, len(rows))
	for i, row := range rows {
		out[i] = toTransaction(row)
	}
	return out
}

func (s *Service) upsertTransactions(ctx context.Context, transactions []plaidlib.Transaction) error {
	upsertErrors := 0
	for _, t := range transactions {
		err := s.upsertTransaction(ctx, t)
		if err != nil {
			log.Printf("[SyncTransactions] failed to upsert transaction %s: %v", t.TransactionId, err)
			upsertErrors++
		}
	}

	if upsertErrors > 0 {
		return fmt.Errorf("failed to upsert %d/%d transactions", upsertErrors, len(transactions))
	}
	return nil
}

func (s *Service) upsertTransaction(ctx context.Context, t plaidlib.Transaction) error {
	datetime := t.Datetime
	var date time.Time
	if datetime.IsSet() {
		date = *datetime.Get()
	} else { // take the date from the transaction obj and fill time 00:00:00
		currTime := time.Now()
		strDate := strings.Split(t.Date, "-")
		year, _ := strconv.Atoi(strDate[0])
		month, _ := strconv.Atoi(strDate[1])
		day, _ := strconv.Atoi(strDate[2])
		date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, currTime.Location())
	}

	params := sqlcdb.UpsertTransactionParams{
		PlaidTransactionID: t.TransactionId,
		PlaidAccountID:     t.AccountId,
		TransactionDate:    pgtype.Date{Valid: true, Time: date},
		TransactionName:    t.Name,
		Amount:             util.Float64ToNumeric(t.GetAmount()),
		Pending:            t.Pending,
		MerchantName:       pgtype.Text{String: t.GetMerchantName(), Valid: t.MerchantName.IsSet()},
		LogoUrl:            pgtype.Text{String: t.GetLogoUrl(), Valid: t.LogoUrl.IsSet()},
		CategoryIconUrl:    pgtype.Text{String: t.GetPersonalFinanceCategoryIconUrl(), Valid: t.HasPersonalFinanceCategoryIconUrl()},
	}

	// Extract personal finance category fields if available
	if pfc, ok := t.GetPersonalFinanceCategoryOk(); ok && pfc != nil {
		params.PersonalFinanceCategory = pgtype.Text{String: pfc.GetPrimary(), Valid: true}
		params.DetailedCategory = pgtype.Text{String: pfc.GetDetailed(), Valid: true}
		params.CategoryConfidenceLevel = pgtype.Text{String: string(pfc.GetConfidenceLevel()), Valid: true}
	}

	_, err := s.repo.Upsert(ctx, params)
	if err != nil {
		log.Printf("[SyncTransactions] failed to upsert transaction %s: %v", t.TransactionId, err)

	}

	return nil
}
