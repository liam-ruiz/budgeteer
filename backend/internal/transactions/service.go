package transactions

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
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
func (s *Service) GetByUser(ctx context.Context, userID uuid.UUID) ([]Transaction, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// CreateTransaction persists a single transaction.
func (s *Service) CreateTransaction(ctx context.Context, params sqlcdb.CreateTransactionParams) (Transaction, error) {
	return s.repo.Create(ctx, params)
}

func (s *Service) CreateTransactions(ctx context.Context, update TransactionUpdate) (error) {
	var params []sqlcdb.CreateTransactionsParams
	// take all of the transactions and add them to the database
	for _, t := range update.Added {
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
		params = append(params, sqlcdb.CreateTransactionsParams{
			PlaidTransactionID: t.TransactionId,
			PlaidAccountID:     t.AccountId,
			TransactionDate:    pgtype.Date{Valid: true, Time: date},
			TransactionName:    t.Name,
			Category:           "Expense", // TODO: implement categorization
			Amount:             floatToNumeric(t.GetAmount()),
			Pending:            t.Pending,
		})
	}
	err := s.repo.CreateMany(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func floatToNumeric(f float64) pgtype.Numeric {
	// convert float64 to bigInt
	bigInt := big.NewInt(int64(f))
	return pgtype.Numeric{Valid: true, Int: bigInt}
}

