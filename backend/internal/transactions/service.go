package transactions

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/util"
	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

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
	return s.repo.GetByUserID(ctx, userID)
}

// CreateTransaction persists a single transaction.
func (s *Service) CreateTransaction(ctx context.Context, params sqlcdb.CreateTransactionParams) (Transaction, error) {
	return s.repo.Create(ctx, params)
}

// CreateTransactions upserts transactions from a Plaid sync update.
// Both Added and Modified transactions are upserted into the database.
func (s *Service) CreateTransactions(ctx context.Context, update TransactionUpdate) error {
	allTransactions := make([]plaidlib.Transaction, 0, len(update.Added)+len(update.Modified))
	allTransactions = append(allTransactions, update.Added...)
	allTransactions = append(allTransactions, update.Modified...)

	if len(allTransactions) == 0 {
		log.Printf("[CreateTransactions] no transactions to upsert for item %s", update.PlaidItemID)
		return nil
	}

	var upsertErrors int
	for _, t := range allTransactions {
		datetime := t.Datetime
		var date time.Time
		if datetime.IsSet() {
			date = *datetime.Get()
		} else { // take the date from the transaction obj and fill time with current time
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
			CategoryIconUrl:    pgtype.Text{String: t.GetPersonalFinanceCategoryIconUrl(), Valid: t.PersonalFinanceCategoryIconUrl != nil},
		}

		// Extract personal finance category fields if available
		if pfc, ok := t.GetPersonalFinanceCategoryOk(); ok && pfc != nil {
			params.PersonalFinanceCategory = pgtype.Text{String: pfc.GetPrimary(), Valid: true}
			params.DetailedCategory = pgtype.Text{String: pfc.GetDetailed(), Valid: true}
			params.CategoryConfidenceLevel = pgtype.Text{String: string(pfc.GetConfidenceLevel()), Valid: true}
		}

		_, err := s.repo.Upsert(ctx, params)
		if err != nil {
			log.Printf("[CreateTransactions] failed to upsert transaction %s: %v", t.TransactionId, err)
			upsertErrors++
		}
	}

	if upsertErrors > 0 {
		return fmt.Errorf("failed to upsert %d/%d transactions", upsertErrors, len(allTransactions))
	}

	log.Printf("[CreateTransactions] successfully upserted %d transactions for item %s", len(allTransactions), update.PlaidItemID)
	return nil
}
