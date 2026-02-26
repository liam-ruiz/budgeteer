package budgets

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/util"
)

// Repository defines the interface for budget data access.
type Repository interface {
	Create(ctx context.Context, params sqlcdb.CreateBudgetParams) (Budget, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetByID(ctx context.Context, id uuid.UUID) (Budget, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new budget repository backed by sqlc queries.
func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, params sqlcdb.CreateBudgetParams) (Budget, error) {
	row, err := r.q.CreateBudget(ctx, params)
	if err != nil {
		return Budget{}, err
	}
	return toBudget(row), nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Budget, error) {
	rows, err := r.q.GetBudgetsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Budget, len(rows))
	for i, row := range rows {
		out[i] = toBudget(row)
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (Budget, error) {
	row, err := r.q.GetBudgetByID(ctx, id)
	if err != nil {
		return Budget{}, err
	}
	return toBudget(row), nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteBudget(ctx, id)
}

func toBudget(row sqlcdb.Budget) Budget {
	var endDate sql.NullTime
	if !row.EndDate.Valid {
		endDate = sql.NullTime{}
	} else {
		endDate = sql.NullTime{
			Time: row.EndDate.Time,
			Valid: true,
		}
	}

	return Budget{
		ID:          row.ID,
		AppUserID:   row.AppUserID,
		Category:    row.Category,
		LimitAmount: util.NumericToString(row.LimitAmount),
		Period:      row.BudgetPeriod,
		StartDate:   row.StartDate.Time,
		EndDate:     endDate,
		CreatedAt:   row.CreatedAt.Time,
	}	
	
}
