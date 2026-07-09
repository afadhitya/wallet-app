-- name: GetBudgetByID :one
SELECT * FROM budgets WHERE id = ?;

-- name: GetBudgetByNameAndPeriod :one
SELECT * FROM budgets
WHERE (name = sqlc.arg('name') OR (name IS NULL AND sqlc.arg('name') IS NULL))
AND period_start = sqlc.arg('period_start')
AND period_end = sqlc.arg('period_end');

-- name: CreateBudget :one
INSERT INTO budgets (name, amount, currency, type, period_start, period_end, notify_at_pct, all_categories)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: ListActiveBudgets :many
SELECT * FROM budgets WHERE is_active = 1 ORDER BY created_at DESC;

-- name: ListAllBudgets :many
SELECT * FROM budgets ORDER BY created_at DESC;

-- name: UpdateBudget :one
UPDATE budgets SET
    name = COALESCE(sqlc.narg('name'), name),
    amount = COALESCE(sqlc.narg('amount'), amount),
    notify_at_pct = COALESCE(sqlc.narg('notify_at_pct'), notify_at_pct),
    period_start = COALESCE(sqlc.narg('period_start'), period_start),
    period_end = COALESCE(sqlc.narg('period_end'), period_end),
    type = COALESCE(sqlc.narg('type'), type),
    all_categories = COALESCE(sqlc.narg('all_categories'), all_categories),
    updated_at = datetime('now')
WHERE id = sqlc.arg('id') RETURNING *;

-- name: MarkBudgetInactive :exec
UPDATE budgets SET is_active = 0, updated_at = datetime('now') WHERE id = ?;

-- name: GetMostRecentPriorBudget :one
SELECT * FROM budgets
WHERE name = sqlc.arg('name')
AND period_end < sqlc.arg('period_end')
ORDER BY period_end DESC
LIMIT 1;

-- name: ListBudgetCategories :many
SELECT c.* FROM categories c
JOIN budget_categories bc ON bc.category_id = c.id
WHERE bc.budget_id = ?;

-- name: ListBudgetTags :many
SELECT t.* FROM tags t
JOIN budget_tags bt ON bt.tag_id = t.id
WHERE bt.budget_id = ?;

-- name: AddBudgetCategory :exec
INSERT INTO budget_categories (budget_id, category_id) VALUES (?, ?);

-- name: AddBudgetTag :exec
INSERT INTO budget_tags (budget_id, tag_id) VALUES (?, ?);

-- name: RemoveBudgetCategory :exec
DELETE FROM budget_categories WHERE budget_id = ? AND category_id = ?;

-- name: RemoveBudgetTag :exec
DELETE FROM budget_tags WHERE budget_id = ? AND tag_id = ?;

-- name: RemoveAllBudgetCategories :exec
DELETE FROM budget_categories WHERE budget_id = ?;

-- name: RemoveAllBudgetTags :exec
DELETE FROM budget_tags WHERE budget_id = ?;

-- name: SumCategoryExpenses :one
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
JOIN budget_categories bc ON bc.category_id = t.category_id
WHERE bc.budget_id = sqlc.arg('budget_id')
AND t.type = 'expense'
AND t.is_archived = 0
AND t.date >= sqlc.arg('period_start')
AND t.date <= sqlc.arg('period_end');

-- name: SumAllCategoryExpenses :one
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE c.type = 'expense'
AND t.type = 'expense'
AND t.is_archived = 0
AND t.date >= sqlc.arg('period_start')
AND t.date <= sqlc.arg('period_end');

-- name: SumTagExpenses :one
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
JOIN transaction_tags tt ON tt.transaction_id = t.id
JOIN budget_tags bt ON bt.tag_id = tt.tag_id
WHERE bt.budget_id = sqlc.arg('budget_id')
AND t.type = 'expense'
AND t.is_archived = 0
AND t.date >= sqlc.arg('period_start')
AND t.date <= sqlc.arg('period_end');
