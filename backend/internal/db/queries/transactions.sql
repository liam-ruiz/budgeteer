-- name: UpsertTransaction :one
INSERT INTO
    transactions (
        plaid_transaction_id,
        plaid_account_id,
        transaction_date,
        transaction_name,
        amount,
        pending,
        merchant_name,
        logo_url,
        personal_finance_category,
        detailed_category,
        category_confidence_level,
        category_icon_url
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
ON CONFLICT (plaid_transaction_id) DO
UPDATE
SET
    transaction_date = EXCLUDED.transaction_date,
    transaction_name = EXCLUDED.transaction_name,
    amount = EXCLUDED.amount,
    pending = EXCLUDED.pending,
    merchant_name = EXCLUDED.merchant_name,
    logo_url = EXCLUDED.logo_url,
    personal_finance_category = EXCLUDED.personal_finance_category,
    detailed_category = EXCLUDED.detailed_category,
    category_confidence_level = EXCLUDED.category_confidence_level,
    category_icon_url = EXCLUDED.category_icon_url
RETURNING
    *;

-- name: CreateTransaction :one
INSERT INTO
    transactions (
        plaid_transaction_id,
        plaid_account_id,
        transaction_date,
        transaction_name,
        amount,
        pending,
        merchant_name,
        logo_url,
        personal_finance_category,
        detailed_category,
        category_confidence_level,
        category_icon_url
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
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
    amount,
    pending,
    merchant_name,
    logo_url,
    personal_finance_category,
    detailed_category,
    category_confidence_level,
    category_icon_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);