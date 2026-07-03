-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = ?;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, category_id, type, amount, currency, description, notes, transfer_to_id, date, is_archived)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0) RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE is_archived = 0
AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'))
AND (sqlc.narg('category_id') IS NULL OR category_id = sqlc.narg('category_id'))
AND (sqlc.narg('type') IS NULL OR type = sqlc.narg('type'))
AND (sqlc.narg('date_from') IS NULL OR date >= sqlc.narg('date_from'))
AND (sqlc.narg('date_to') IS NULL OR date <= sqlc.narg('date_to'))
ORDER BY date DESC, id DESC
LIMIT ? OFFSET ?;

-- name: ListTransactionsByTag :many
SELECT t.* FROM transactions t
JOIN transaction_tags tt ON tt.transaction_id = t.id
JOIN tags ON tags.id = tt.tag_id
WHERE t.is_archived = 0
AND tags.name = sqlc.arg('tag_name') COLLATE NOCASE
AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
AND (sqlc.narg('category_id') IS NULL OR t.category_id = sqlc.narg('category_id'))
AND (sqlc.narg('type') IS NULL OR t.type = sqlc.narg('type'))
AND (sqlc.narg('date_from') IS NULL OR t.date >= sqlc.narg('date_from'))
AND (sqlc.narg('date_to') IS NULL OR t.date <= sqlc.narg('date_to'))
ORDER BY t.date DESC, t.id DESC
LIMIT ? OFFSET ?;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE is_archived = 0
AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'))
AND (sqlc.narg('category_id') IS NULL OR category_id = sqlc.narg('category_id'))
AND (sqlc.narg('type') IS NULL OR type = sqlc.narg('type'))
AND (sqlc.narg('date_from') IS NULL OR date >= sqlc.narg('date_from'))
AND (sqlc.narg('date_to') IS NULL OR date <= sqlc.narg('date_to'));

-- name: SumTransactions :one
SELECT COALESCE(SUM(amount), 0) FROM transactions
WHERE is_archived = 0
AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'))
AND (sqlc.narg('category_id') IS NULL OR category_id = sqlc.narg('category_id'))
AND (sqlc.narg('type') IS NULL OR type = sqlc.narg('type'))
AND (sqlc.narg('date_from') IS NULL OR date >= sqlc.narg('date_from'))
AND (sqlc.narg('date_to') IS NULL OR date <= sqlc.narg('date_to'));

-- name: UpdateTransaction :one
UPDATE transactions SET
    amount = COALESCE(sqlc.narg('amount'), amount),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    account_id = COALESCE(sqlc.narg('account_id'), account_id),
    date = COALESCE(sqlc.narg('date'), date),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    updated_at = datetime('now')
WHERE id = sqlc.arg('id') RETURNING *;

-- name: ArchiveTransaction :exec
UPDATE transactions SET is_archived = 1, updated_at = datetime('now') WHERE id = ?;

-- name: GetAccountBalance :one
SELECT COALESCE(SUM(
    CASE
        WHEN t.type = 'expense' THEN -t.amount
        WHEN t.type = 'transfer' AND t.account_id = sqlc.arg('account_id') THEN -t.amount
        WHEN t.type = 'transfer' AND t.transfer_to_id = sqlc.arg('account_id') THEN t.amount
        ELSE t.amount
    END
), 0) AS balance
FROM transactions t
WHERE t.is_archived = 0
AND (
    t.account_id = sqlc.arg('account_id')
    OR (t.type = 'transfer' AND t.transfer_to_id = sqlc.arg('account_id'))
);

-- name: GetDefaultAccount :one
SELECT * FROM accounts WHERE is_archived = 0 ORDER BY sort_order, id LIMIT 1;
