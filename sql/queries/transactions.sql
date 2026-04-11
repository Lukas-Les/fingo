-- name: CreateTransaction :one
INSERT INTO transactions (id, created_at, updated_at, user_id, amount, transaction_type, category, description, party, transaction_date)
VALUES (
    gen_random_uuid(),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: GetUserTransactions :many
SELECT * FROM transactions WHERE user_id = $1;

-- name: GetUserTransactionsForPeriod :many
SELECT * FROM transactions WHERE user_id = $1 and transaction_date <= $2 and transaction_date >= $3;
