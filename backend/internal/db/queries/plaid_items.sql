-- name: CreatePlaidItem :one
INSERT INTO
    plaid_items (
        app_user_id,
        plaid_item_id,
        plaid_access_token,
        institution_name
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetPlaidItemsByUserID :many
SELECT *
FROM plaid_items
WHERE
    app_user_id = $1
ORDER BY institution_name;

-- name: GetPlaidItemByID :one
SELECT * FROM plaid_items WHERE plaid_item_id = $1;

-- name: GetPlaidItemByPlaidItemID :one
SELECT * FROM plaid_items WHERE plaid_item_id = $1 LIMIT 1;

-- name: UpdatePlaidItemCursor :exec
UPDATE plaid_items
SET
    plaid_cursor = $2,
    last_synced_at = now(),
    updated_at = now()
WHERE
    plaid_item_id = $1;

-- name: GetCursor :one
SELECT plaid_cursor FROM plaid_items WHERE plaid_item_id = $1 LIMIT 1;
