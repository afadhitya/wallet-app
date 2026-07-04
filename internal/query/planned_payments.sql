-- name: GetPlannedPaymentByID :one
SELECT * FROM planned_payments WHERE id = ?;

-- name: ListActivePlannedPayments :many
SELECT * FROM planned_payments
WHERE is_active = 1 AND is_paused = 0
ORDER BY COALESCE(next_due_date, start_date) ASC;

-- name: ListPausedPlannedPayments :many
SELECT * FROM planned_payments
WHERE is_active = 1 AND is_paused = 1
ORDER BY COALESCE(next_due_date, start_date) ASC;

-- name: ListArchivedPlannedPayments :many
SELECT * FROM planned_payments
WHERE is_active = 0
ORDER BY updated_at DESC;

-- name: ListAllPlannedPayments :many
SELECT * FROM planned_payments
ORDER BY COALESCE(next_due_date, start_date) ASC;

-- name: ListDuePlannedPayments :many
SELECT * FROM planned_payments
WHERE is_active = 1 AND is_paused = 0
AND next_due_date IS NOT NULL
AND next_due_date >= sqlc.arg('date_from')
AND next_due_date <= sqlc.arg('date_to')
ORDER BY next_due_date ASC;

-- name: ListOverduePlannedPayments :many
SELECT * FROM planned_payments
WHERE is_active = 1 AND is_paused = 0
AND next_due_date IS NOT NULL
AND next_due_date < sqlc.arg('today')
ORDER BY next_due_date ASC;

-- name: CreatePlannedPayment :one
INSERT INTO planned_payments (
    account_id, category_id, type, amount, currency, name,
    recurrence, recurrence_rule, start_date, next_due_date,
    is_paused, is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1) RETURNING *;

-- name: UpdatePlannedPayment :one
UPDATE planned_payments SET
    name = COALESCE(sqlc.narg('name'), name),
    amount = COALESCE(sqlc.narg('amount'), amount),
    currency = COALESCE(sqlc.narg('currency'), currency),
    type = COALESCE(sqlc.narg('type'), type),
    account_id = COALESCE(sqlc.narg('account_id'), account_id),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    recurrence = COALESCE(sqlc.narg('recurrence'), recurrence),
    recurrence_rule = COALESCE(sqlc.narg('recurrence_rule'), recurrence_rule),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    next_due_date = COALESCE(sqlc.narg('next_due_date'), next_due_date),
    updated_at = datetime('now')
WHERE id = sqlc.arg('id') RETURNING *;

-- name: UpdatePlannedPaymentNextDueDate :one
UPDATE planned_payments SET
    next_due_date = sqlc.arg('next_due_date'),
    updated_at = datetime('now')
WHERE id = sqlc.arg('id') RETURNING *;

-- name: ArchivePlannedPayment :exec
UPDATE planned_payments SET is_active = 0, updated_at = datetime('now') WHERE id = ?;

-- name: PausePlannedPayment :exec
UPDATE planned_payments SET is_paused = 1, updated_at = datetime('now') WHERE id = ?;

-- name: ResumePlannedPayment :exec
UPDATE planned_payments SET is_paused = 0, updated_at = datetime('now') WHERE id = ?;

-- name: ListActivePlannedPaymentsForAccount :many
SELECT * FROM planned_payments
WHERE is_active = 1 AND is_paused = 0
AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'))
ORDER BY COALESCE(next_due_date, start_date) ASC;

-- name: DeletePlannedPayment :exec
DELETE FROM planned_payments WHERE id = ?;
