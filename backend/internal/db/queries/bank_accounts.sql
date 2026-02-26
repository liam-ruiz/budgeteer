-- name: CreateBankAccount :one
INSERT INTO
    bank_accounts (
        plaid_account_id,
        plaid_item_id,
        account_name,
        official_name,
        account_type,
        account_subtype,
        current_balance,
        available_balance,
        iso_currency_code
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
        $9
    )
RETURNING
    *;


-- name: UpsertBankAccount :one
INSERT INTO
    bank_accounts (
        plaid_account_id,
        plaid_item_id,
        account_name,
        official_name,
        account_type,
        account_subtype,
        current_balance,
        available_balance,
        iso_currency_code
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
        $9
    )
ON CONFLICT (plaid_account_id) DO UPDATE
    SET
        plaid_item_id = $2,
        account_name = $3,
        official_name = $4,
        account_type = $5,
        account_subtype = $6,
        current_balance = $7,
        available_balance = $8,
        iso_currency_code = $9,
        updated_at = now()
RETURNING
    *;

-- name: GetBankAccountsByItemID :many
SELECT *
FROM bank_accounts
WHERE
    plaid_item_id = $1
ORDER BY account_name;

-- name: GetBankAccountsByUserID :many
SELECT ba.*
FROM
    bank_accounts ba
    JOIN plaid_items pli ON ba.plaid_item_id = pli.plaid_item_id
WHERE
    pli.app_user_id = $1
ORDER BY ba.account_name;

-- name: GetBankAccountByPlaidAccountID :one
SELECT * FROM bank_accounts WHERE plaid_account_id = $1 LIMIT 1;

-- name: UpdateBankAccountBalance :exec
UPDATE bank_accounts
SET
    current_balance = $2,
    available_balance = $3,
    updated_at = now()
WHERE
    plaid_account_id = $1;