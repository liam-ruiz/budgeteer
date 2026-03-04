package budgets

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Budget represents a spending budget for a category.
type Budget struct {
	ID          uuid.UUID    `json:"id"`
	AppUserID   uuid.UUID    `json:"user_id"`
	Category    string       `json:"category"`
	LimitAmount string       `json:"limit_amount"`
	AmountSpent string       `json:"amount_spent"`
	Period      string       `json:"period"`
	StartDate   time.Time    `json:"start_date"`
	EndDate     sql.NullTime `json:"end_date"`
	CreatedAt   time.Time    `json:"created_at"`
}

// BudgetResponse is the JSON-safe view sent to clients.
type BudgetResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Category    string    `json:"category"`
	LimitAmount string    `json:"limit_amount"`
	AmountSpent string    `json:"amount_spent"`
	Period      string    `json:"period"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToBudgetResponse(b Budget) BudgetResponse {
	var end *string
	if b.EndDate.Valid {
		s := b.EndDate.Time.Format("2006-01-02")
		end = &s
	}
	return BudgetResponse{
		ID:          b.ID,
		UserID:      b.AppUserID,
		Category:    b.Category,
		LimitAmount: b.LimitAmount,
		AmountSpent: b.AmountSpent,
		Period:      b.Period,
		StartDate:   b.StartDate.Format("2006-01-02"),
		EndDate:     end,
		CreatedAt:   b.CreatedAt,
	}
}
