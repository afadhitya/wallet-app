-- name: ReportIncomeByCategory :many
SELECT
    c.id AS category_id,
    c.name AS category_name,
    SUM(COALESCE(t.base_amount, t.amount)) AS total,
    COUNT(*) AS transaction_count
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.is_archived = 0
  AND t.type = 'income'
  AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
GROUP BY c.id
ORDER BY total DESC;

-- name: ReportExpenseByCategory :many
SELECT
    c.id AS category_id,
    c.name AS category_name,
    pc.name AS parent_category_name,
    SUM(COALESCE(t.base_amount, t.amount)) AS total,
    COUNT(*) AS transaction_count
FROM transactions t
JOIN categories c ON t.category_id = c.id
LEFT JOIN categories pc ON c.parent_id = pc.id
WHERE t.is_archived = 0
  AND t.type = 'expense'
  AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
GROUP BY c.id
ORDER BY total DESC;

-- name: ReportByAccount :many
SELECT
    a.id AS account_id,
    a.name AS account_name,
    a.currency,
    COALESCE(SUM(CASE WHEN t.type = 'income' THEN COALESCE(t.base_amount, t.amount) END), 0) AS income,
    COALESCE(SUM(CASE WHEN t.type = 'expense' THEN COALESCE(t.base_amount, t.amount) END), 0) AS expense
FROM accounts a
LEFT JOIN transactions t ON t.account_id = a.id
    AND t.is_archived = 0
    AND t.type IN ('income', 'expense')
    AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
WHERE (sqlc.narg('account_id') IS NULL OR a.id = sqlc.narg('account_id'))
GROUP BY a.id
HAVING income > 0 OR expense > 0
ORDER BY income DESC, expense DESC;

-- name: ReportByTag :many
SELECT
    tg.id AS tag_id,
    tg.name AS tag_name,
    SUM(COALESCE(t.base_amount, t.amount)) AS total,
    COUNT(*) AS transaction_count
FROM transactions t
JOIN transaction_tags tt ON tt.transaction_id = t.id
JOIN tags tg ON tt.tag_id = tg.id
WHERE t.is_archived = 0
  AND t.type = 'expense'
  AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
GROUP BY tg.id
ORDER BY total DESC;

-- name: ReportUntagged :one
SELECT
    COALESCE(SUM(COALESCE(t.base_amount, t.amount)), 0) AS total,
    COUNT(*) AS transaction_count
FROM transactions t
WHERE t.is_archived = 0
  AND t.type = 'expense'
  AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
  AND NOT EXISTS (SELECT 1 FROM transaction_tags tt WHERE tt.transaction_id = t.id);

-- name: ReportTransfers :one
SELECT
    COALESCE(SUM(COALESCE(base_amount, amount)), 0) AS total
FROM transactions
WHERE is_archived = 0
  AND type = 'transfer'
  AND date >= sqlc.arg('date_from') AND date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'));

-- name: ReportTransactionCount :one
SELECT
    COUNT(*) AS total
FROM transactions
WHERE is_archived = 0
  AND type != 'adjustment'
  AND date >= sqlc.arg('date_from') AND date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'));

-- name: ReportExpenseTotal :one
SELECT
    COALESCE(SUM(COALESCE(base_amount, amount)), 0) AS total
FROM transactions
WHERE is_archived = 0
  AND type = 'expense'
  AND date >= sqlc.arg('date_from') AND date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'));

-- name: ReportIncomeTotal :one
SELECT
    COALESCE(SUM(COALESCE(base_amount, amount)), 0) AS total
FROM transactions
WHERE is_archived = 0
  AND type = 'income'
  AND date >= sqlc.arg('date_from') AND date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR account_id = sqlc.narg('account_id'));

-- name: ReportExportTransactions :many
SELECT
    t.id,
    t.date,
    t.type,
    t.amount,
    t.currency,
    t.base_amount,
    c.name AS category_name,
    a.name AS account_name,
    t.description
FROM transactions t
JOIN accounts a ON t.account_id = a.id
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.is_archived = 0
  AND t.type != 'adjustment'
  AND t.date >= sqlc.arg('date_from') AND t.date <= sqlc.arg('date_to')
  AND (sqlc.narg('account_id') IS NULL OR t.account_id = sqlc.narg('account_id'))
ORDER BY t.date DESC, t.id DESC;
