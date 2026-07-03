-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = ?;

-- name: GetAccountByName :one
SELECT * FROM accounts WHERE name = ? COLLATE NOCASE;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE is_archived = 0 ORDER BY sort_order, name;

-- name: ListAllAccounts :many
SELECT * FROM accounts ORDER BY sort_order, name;

-- name: CreateAccount :one
INSERT INTO accounts (name, type, currency) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts SET name = ?, type = ?, currency = ?, updated_at = datetime('now') WHERE id = ? RETURNING *;

-- name: UpdateAccountBalance :exec
UPDATE accounts SET balance = ?, updated_at = datetime('now') WHERE id = ?;

-- name: ArchiveAccount :exec
UPDATE accounts SET is_archived = 1, updated_at = datetime('now') WHERE id = ?;
