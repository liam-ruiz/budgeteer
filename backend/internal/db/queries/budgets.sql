-- name: CreateBudget :one
INSERT INTO
    budgets (
        app_user_id,
        category,
        limit_amount,
        budget_period,
        start_date,
        end_date
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetBudgetsByUserID :many
SELECT * FROM budgets WHERE app_user_id = $1 ORDER BY category;

-- name: GetBudgetByID :one
SELECT * FROM budgets WHERE id = $1;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1;

-- name: UpdateBudgetAmountSpent :one
UPDATE budgets SET amount_spent = $2 WHERE id = $1 RETURNING *;