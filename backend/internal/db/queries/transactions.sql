-- name: UpsertTransaction :one
INSERT INTO
    transactions (
        plaid_transaction_id,
        plaid_account_id,
        transaction_date,
        transaction_name,
        category,
        amount,
        pending
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (plaid_transaction_id) DO
UPDATE
SET
    transaction_date = EXCLUDED.transaction_date,
    transaction_name = EXCLUDED.transaction_name,
    category = EXCLUDED.category,
    amount = EXCLUDED.amount,
    pending = EXCLUDED.pending
RETURNING
    *;

-- name: CreateTransaction :one
INSERT INTO
    transactions (
        plaid_transaction_id,
        plaid_account_id,
        transaction_date,
        transaction_name,
        category,
        amount,
        pending
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetTransactionsByAccountID :many
SELECT *
FROM transactions
WHERE
    plaid_account_id = $1
ORDER BY transaction_date DESC;

-- name: GetTransactionsByUserID :many
SELECT t.*
FROM
    transactions t
    JOIN bank_accounts ba ON t.plaid_account_id = ba.plaid_account_id
    JOIN plaid_items pli ON ba.plaid_item_id = pli.plaid_item_id
WHERE
    pli.app_user_id = $1
ORDER BY t.transaction_date DESC;

-- name: CreateTransactions :copyfrom
INSERT INTO transactions (
    plaid_transaction_id,
    plaid_account_id,
    transaction_date,
    transaction_name,
    category,
    amount,
    pending
) VALUES ($1, $2, $3, $4, $5, $6, $7);