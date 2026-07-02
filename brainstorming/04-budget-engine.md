# 04 — Budget Engine

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md)
> Status: 🔴 pending review | Unblocks: 05-planned-payments

---

## Objective

Implement budget management — set spending limits per category/tag, check progress, and auto-generate recurring periods.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| B1 | Budget format | Named — multi-target via flags | Flexible, clear identifier |
| B2 | Budget checking | On-demand (`wallet budget check`) | MVP simple, no background watcher |
| B3 | Period auto-creation | Auto-generate on first check in new month | Seamless recurring budgets |

---

## Commands

### `wallet budget set`

Create or update a budget.

```
$ wallet budget set "Monthly Food" 2000000 -c food -c transport --period monthly
✓ Budget created: "Monthly Food" — Rp2.000.000/month
  Targets: Food & Dining, Transportation
  Period: 2026-07-01 → 2026-07-31
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--category` | `-c` | No* | — | Category name or ID (repeatable) |
| `--tag` | `-t` | No* | — | Tag name (repeatable) |
| `--period` | | No | monthly | monthly / weekly / yearly / one_time |
| `--from` | | No | auto | Period start date (YYYY-MM-DD) |
| `--to` | | No | auto | Period end date (YYYY-MM-DD) |
| `--notify` | | No | 80 | Alert at X% spent |
| `--json` | | No | false | JSON output |

*At least one `-c` or `-t` required. Can mix? No — either categories OR tags, enforced at app layer.

**Validation:**
- At least one target (category or tag) required
- Amount must be > 0
- Period `one_time` requires `--from` and `--to`
- Period `monthly` auto-sets from/to to current month
- If budget name exists, update instead of create (upsert)

**Period auto-calculation:**

| Period | from | to |
|--------|------|----|
| monthly | 1st of current month | last day of current month |
| weekly | monday of current week | sunday of current week |
| yearly | jan 1 of current year | dec 31 of current year |
| one_time | user-specified | user-specified |

---

### `wallet budget list`

List all active budgets.

```
$ wallet budget list
┌────┬──────────────────┬────────────┬─────────────┬────────────┐
│ ID │ Name             │ Limit      │ Spent       │ Remaining  │
├────┼──────────────────┼────────────┼─────────────┼────────────┤
│  1 │ Monthly Food     │ Rp2.000.000│ Rp1.250.000 │ Rp750.000  │
│  2 │ Japan Trip 2026  │ Rp10.000.000│ Rp3.500.000│ Rp6.500.000│
└────┴──────────────────┴────────────┴─────────────┴────────────┘
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--active` | Show only active budgets (default) |
| `--all` | Include inactive/expired |
| `--json` | JSON output |

---

### `wallet budget check`

Check budget progress — shows spending vs limit for current period.

```
$ wallet budget check --all
┌──────────────────┬────────────┬─────────────┬─────────┬─────────────────────────┐
│ Budget           │ Limit      │ Spent       │ %       │ Status                  │
├──────────────────┼────────────┼─────────────┼─────────┼─────────────────────────┤
│ Monthly Food     │ Rp2.000.000│ Rp1.800.000 │ 90.0%   │ ⚠️ WARNING (90%)        │
│ Japan Trip 2026  │ Rp10.000.000│ Rp3.500.000 │ 35.0%   │ ✅ OK                   │
│ Bills            │ Rp500.000  │ Rp500.000   │ 100.0%  │ 🔴 OVER BUDGET          │
└──────────────────┴────────────┴─────────────┴─────────┴─────────────────────────┘
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--budget` | `-b` | Check specific budget by name or ID |
| `--all` | | Check all active budgets |
| `--json` | | JSON output |

**Status logic:**

| % Spent | Status |
|---------|--------|
| < notify_at_pct | ✅ OK |
| >= notify_at_pct and < 100% | ⚠️ WARNING |
| >= 100% | 🔴 OVER BUDGET |

**Auto-generation trigger:**
- If a recurring budget has no period for current month, auto-create one
- New period copies limit from previous period
- `period_start` = 1st of current month, `period_end` = last day

---

### `wallet budget edit <id>`

Edit an existing budget.

```
$ wallet budget edit 1 --amount 2500000 --notify 75
✓ Updated budget #1: "Monthly Food" — Rp2.500.000/month (alert at 75%)
```

**Editable fields:**
- `--amount` — limit
- `--name` — budget name
- `--notify` — alert percentage
- `--add-category` / `--remove-category`
- `--add-tag` / `--remove-tag`

---

### `wallet budget rm <id>`

Delete a budget.

```
$ wallet budget rm 1
✓ Deleted budget #1: "Monthly Food"
```

---

## Service Layer

### BudgetService

```go
type BudgetService struct {
    db *gen.Queries
}

func (s *BudgetService) Set(ctx, params) (*Budget, error)
func (s *BudgetService) List(ctx, activeOnly) ([]Budget, error)
func (s *BudgetService) Check(ctx, budgetID or all) ([]BudgetStatus, error)
func (s *BudgetService) Edit(ctx, id, params) (*Budget, error)
func (s *BudgetService) Delete(ctx, id) error
func (s *BudgetService) AutoGeneratePeriod(ctx, budgetID) error
func (s *BudgetService) CalcSpent(ctx, budgetID) (int64, error)
```

### BudgetStatus (response struct)

```go
type BudgetStatus struct {
    BudgetID    int64
    Name        string
    Limit       int64
    Spent       int64
    Remaining   int64
    PercentUsed float64
    Status      string // "ok", "warning", "over"
}
```

---

## sqlc Queries

### budgets.sql

```sql
-- name: GetBudget :one
SELECT * FROM budgets WHERE id = ?;

-- name: ListBudgets :many
SELECT * FROM budgets WHERE is_active = 1 ORDER BY created_at DESC;

-- name: CreateBudget :one
INSERT INTO budgets (name, amount, currency, type, period_start, period_end, notify_at_pct, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, 1) RETURNING *;

-- name: UpdateBudget :exec
UPDATE budgets SET name = ?, amount = ?, notify_at_pct = ?, updated_at = datetime('now') WHERE id = ?;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = ?;

-- name: GetBudgetCategories :many
SELECT c.* FROM categories c
JOIN budget_categories bc ON c.id = bc.category_id
WHERE bc.budget_id = ?;

-- name: GetBudgetTags :many
SELECT t.* FROM tags t
JOIN budget_tags bt ON t.id = bt.tag_id
WHERE bt.budget_id = ?;

-- name: AddBudgetCategory :exec
INSERT OR IGNORE INTO budget_categories (budget_id, category_id) VALUES (?, ?);

-- name: RemoveBudgetCategory :exec
DELETE FROM budget_categories WHERE budget_id = ? AND category_id = ?;

-- name: AddBudgetTag :exec
INSERT OR IGNORE INTO budget_tags (budget_id, tag_id) VALUES (?, ?);

-- name: RemoveBudgetTag :exec
DELETE FROM budget_tags WHERE budget_id = ? AND tag_id = ?;

-- name: SumSpentByCategories :one
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
WHERE t.category_id IN (SELECT category_id FROM budget_categories WHERE budget_id = ?)
  AND t.type = 'expense'
  AND t.date >= ? AND t.date <= ?;

-- name: SumSpentByTags :one
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
JOIN transaction_tags tt ON t.id = tt.transaction_id
WHERE tt.tag_id IN (SELECT tag_id FROM budget_tags WHERE budget_id = ?)
  AND t.type = 'expense'
  AND t.date >= ? AND t.date <= ?;

-- name: FindBudgetByPeriod :one
SELECT * FROM budgets
WHERE name = ? AND period_start = ? AND period_end = ?;
```

---

## Auto-Generation Logic

```
BudgetService.Check(ctx, budgetID):
  1. Load budget
  2. If type == "recurring":
     a. Get current month period (period_start, period_end)
     b. If no period exists for current month:
        - Find most recent past period
        - Create new period with same limit
        - Return new period
  3. Calculate spent (SumSpentByCategories or SumSpentByTags)
  4. Return BudgetStatus
```

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| No targets | `Budget must have at least one category or tag.` | 1 |
| Budget not found | `Budget #99 not found.` | 1 |
| Invalid amount | `Amount must be a positive number.` | 1 |
| Invalid period | `Period 'foo' not supported. Use: monthly, weekly, yearly, one_time` | 1 |
| one_time without dates | `one_time budget requires --from and --to` | 1 |

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Can a budget mix categories AND tags? | → TBD |
| OQ2 | What happens when a budget period overlaps? | → TBD |
