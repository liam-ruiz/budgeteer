package budgets

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
	"github.com/liam-ruiz/budgeteer/internal/util"
)

var ErrBudgetNotFound = errors.New("budget not found")

// Service handles budget business logic.
type Service struct {
	repo Repository
}

// NewService creates a new budget service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateBudget creates a new budget for a user, then calculates its initial spend.
func (s *Service) CreateBudget(ctx context.Context, params sqlcdb.CreateBudgetParams) (BudgetResponse, error) {
	b, err := s.repo.Create(ctx, params)
	if err != nil {
		return BudgetResponse{}, err
	}

	b = s.recalculateSpend(ctx, b)
	return ToBudgetResponse(b), nil
}

// GetBudgets returns all budgets for a user with calculated spend amounts.
func (s *Service) GetBudgets(ctx context.Context, userID uuid.UUID) ([]BudgetResponse, error) {
	budgets, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]BudgetResponse, len(budgets))
	for i, b := range budgets {
		b = s.recalculateSpend(ctx, b)
		out[i] = ToBudgetResponse(b)
	}
	return out, nil
}

// GetBudget returns a single budget owned by the user with recalculated spend.
func (s *Service) GetBudget(ctx context.Context, userID, id uuid.UUID) (BudgetResponse, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return BudgetResponse{}, err
	}
	if b.AppUserID != userID {
		return BudgetResponse{}, ErrBudgetNotFound
	}

	b = s.recalculateSpend(ctx, b)
	return ToBudgetResponse(b), nil
}

// UpdateBudget updates a budget and recalculates its spend.
func (s *Service) UpdateBudget(ctx context.Context, params sqlcdb.UpdateBudgetParams) (BudgetResponse, error) {
	b, err := s.repo.Update(ctx, params)
	if err != nil {
		return BudgetResponse{}, err
	}
	b = s.recalculateSpend(ctx, b)
	return ToBudgetResponse(b), nil
}

// DeleteBudget removes a budget by ID.
func (s *Service) DeleteBudget(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// recalculateSpend computes and updates the amount_spent for a budget.
// If the budget has a category, only transactions matching that category are summed.
// If no category, all transactions within the date range are summed.
func (s *Service) recalculateSpend(ctx context.Context, b Budget) Budget {
	var spent pgtype.Numeric
	var err error

	startDate := pgtype.Date{Time: b.StartDate, Valid: true}
	var endDate pgtype.Date
	if b.EndDate.Valid {
		endDate = pgtype.Date{Time: b.EndDate.Time, Valid: true}
	}

	if b.Category != nil && *b.Category != "" {
		spent, err = s.repo.CalculateSpendByCategory(ctx, sqlcdb.CalculateBudgetSpendByCategoryParams{
			AppUserID:         b.AppUserID,
			Upper:             *b.Category,
			TransactionDate:   startDate,
			TransactionDate_2: endDate,
		})
	} else {
		spent, err = s.repo.CalculateSpendAll(ctx, sqlcdb.CalculateBudgetSpendAllParams{
			AppUserID:         b.AppUserID,
			TransactionDate:   startDate,
			TransactionDate_2: endDate,
		})
	}

	if err != nil {
		log.Printf("[recalculateSpend] error calculating spend for budget %s: %v", b.ID, err)
		return b
	}

	b.AmountSpent = util.NumericToString(spent)
	return b
}
