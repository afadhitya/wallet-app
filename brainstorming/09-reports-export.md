# 09 — Reports & Export

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md), [07-multi-currency](./07-multi-currency.md)
> Status: 🔴 pending review | Unblocks: none (final phase)

---

## Objective

Implement financial reports with flexible breakdown options and CSV export. TUI skipped per Phase 02 decision.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| R1 | Report types | Monthly + breakdown by category/account/tag | Flexible, one command |
| R2 | Export | CSV only | MVP, universal format |

---

## Commands

### `wallet report`

Monthly financial summary.

```
$ wallet report --month july
┌─────────────────────────────────────────────────────────────┐
│ Monthly Report — July 2026                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Income:       Rp15.000.000                                  │
│   Salary:     Rp15.000.000                                  │
│                                                             │
│ Expenses:     Rp8.500.000                                   │
│   Food:       Rp2.500.000                                   │
│   Transport:  Rp1.200.000                                   │
│   Travel:     Rp1.357.000                                   │
│   Bills:      Rp3.443.000                                   │
│                                                             │
│ Net:          +Rp6.500.000                                  │
│                                                             │
│ Transfers:    Rp2.000.000 (between own accounts)            │
└─────────────────────────────────────────────────────────────┘
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--month` | `-m` | current | Month name or YYYY-MM |
| `--from` | | — | Start date (YYYY-MM-DD) |
| `--to` | | — | End date (YYYY-MM-DD) |
| `--account` | `-a` | all | Specific account |
| `--json` | | false | JSON output |
| `--export` | | — | Export to file (csv) |

---

### `wallet report --by category`

Breakdown by category.

```
$ wallet report --month july --by category
┌─────────────────────────────────────────────────────────────┐
│ Expenses by Category — July 2026                            │
├─────────────────────┬─────────────┬─────────┬───────────────┤
│ Category            │ Amount      │ %       │ Transactions  │
├─────────────────────┼─────────────┼─────────┼───────────────┤
│ Food & Dining       │ Rp2.500.000 │ 29.4%   │ 45            │
│   Restaurant        │ Rp1.800.000 │ 21.2%   │ 30            │
│   Groceries         │ Rp500.000   │ 5.9%    │ 10            │
│   Coffee & Snacks   │ Rp200.000   │ 2.4%    │ 5             │
│ Bills & Utilities   │ Rp3.443.000 │ 40.5%   │ 5             │
│ Transportation      │ Rp1.200.000 │ 14.1%   │ 20            │
│ Travel              │ Rp1.357.000 │ 16.0%   │ 3             │
├─────────────────────┼─────────────┼─────────┼───────────────┤
│ TOTAL               │ Rp8.500.000 │ 100%    │ 73            │
└─────────────────────┴─────────────┴─────────┴───────────────┘
```

---

### `wallet report --by account`

Breakdown by account.

```
$ wallet report --month july --by account
┌─────────────────────────────────────────────────────────────┐
│ Transactions by Account — July 2026                         │
├─────────────────┬───────────────┬───────────────┬───────────┤
│ Account         │ Income        │ Expenses      │ Net       │
├─────────────────┼───────────────┼───────────────┼───────────┤
│ BCA             │ Rp15.000.000  │ Rp7.000.000   │+Rp8.000.00│
│ GoPay           │ Rp0           │ Rp1.500.000   │-Rp1.500.00│
│ DBS SGD         │ Rp0           │ Rp1.180.000   │-Rp1.180.00│
├─────────────────┼───────────────┼───────────────┼───────────┤
│ TOTAL           │ Rp15.000.000  │ Rp9.680.000   │+Rp5.320.00│
└─────────────────┴───────────────┴───────────────┴───────────┘
```

---

### `wallet report --by tag`

Breakdown by tag.

```
$ wallet report --month july --by tag
┌─────────────────────────────────────────────────────────────┐
│ Expenses by Tag — July 2026                                 │
├─────────────────────┬─────────────┬─────────┬───────────────┤
│ Tag                 │ Amount      │ %       │ Transactions  │
├─────────────────────┼─────────────┼─────────┼───────────────┤
│ work                │ Rp1.500.000 │ 17.6%   │ 25            │
│ vacation            │ Rp2.000.000 │ 23.5%   │ 10            │
│ reimbursable        │ Rp800.000   │ 9.4%    │ 5             │
│ (untagged)          │ Rp4.200.000 │ 49.4%   │ 33            │
├─────────────────────┼─────────────┼─────────┼───────────────┤
│ TOTAL               │ Rp8.500.000 │ 100%    │ 73            │
└─────────────────────┴─────────────┴─────────┴───────────────┘
```

---

### `wallet report --export csv`

Export report to CSV file.

```
$ wallet report --month july --by category --export csv
✓ Exported to: wallet-report-2026-07.csv
```

**CSV format:**

```csv
date,type,amount,currency,base_amount,category,account,description,tags
2026-07-01,expense,35000,IDR,,Restaurant,BCA,Lunch at Warung,"lunch,work"
2026-07-02,expense,100,SGD,1180000,Travel,DBS SGD,Coffee in Singapore,travel
2026-07-01,income,5000000,IDR,,Salary,BCA,Gaji Juli,
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--export csv` | Export to CSV file |
| `--output` | Custom output filename |

---

### `wallet report --json`

JSON report output.

```json
{
  "success": true,
  "data": {
    "period": "2026-07",
    "income": {
      "total": 15000000,
      "by_category": [
        {"category": "Salary", "amount": 15000000, "percent": 100}
      ]
    },
    "expenses": {
      "total": 8500000,
      "by_category": [
        {"category": "Food & Dining", "amount": 2500000, "percent": 29.4},
        {"category": "Bills & Utilities", "amount": 3443000, "percent": 40.5}
      ]
    },
    "net": 6500000,
    "transfers": 2000000,
    "transaction_count": 73
  }
}
```

---

## Service Layer

### ReportService

```go
type ReportService struct {
    db *gen.Queries
}

type ReportResult struct {
    Period          string
    Income          CategoryBreakdown
    Expenses        CategoryBreakdown
    Net             int64
    Transfers       int64
    TransactionCount int
}

type CategoryBreakdown struct {
    Total     int64
    ByCategory []BreakdownItem
}

type BreakdownItem struct {
    Name       string
    Amount     int64
    Percent    float64
    Count      int
}

func (s *ReportService) Monthly(ctx, month, accountID) (*ReportResult, error)
func (s *ReportService) ByCategory(ctx, filters) ([]BreakdownItem, error)
func (s *ReportService) ByAccount(ctx, filters) ([]BreakdownItem, error)
func (s *ReportService) ByTag(ctx, filters) ([]BreakdownItem, error)
func (s *ReportService) ExportCSV(ctx, filters, outputPath) error
```

---

## sqlc Queries

### reports.sql

```sql
-- name: ReportIncomeByCategory :many
SELECT
    c.name as category,
    SUM(t.amount) as total,
    COUNT(*) as count
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.type = 'income'
  AND t.date >= ? AND t.date <= ?
  AND (? = 0 OR t.account_id = ?)
GROUP BY c.id
ORDER BY total DESC;

-- name: ReportExpenseByCategory :many
SELECT
    c.name as category,
    COALESCE(pc.name, c.name) as parent_category,
    SUM(t.amount) as total,
    COUNT(*) as count
FROM transactions t
JOIN categories c ON t.category_id = c.id
LEFT JOIN categories pc ON c.parent_id = pc.id
WHERE t.type = 'expense'
  AND t.date >= ? AND t.date <= ?
  AND (? = 0 OR t.account_id = ?)
GROUP BY c.id, pc.id
ORDER BY total DESC;

-- name: ReportByAccount :many
SELECT
    a.name as account,
    a.currency,
    SUM(CASE WHEN t.type = 'income' THEN t.base_amount ELSE 0 END) as income,
    SUM(CASE WHEN t.type = 'expense' THEN t.base_amount ELSE 0 END) as expenses
FROM transactions t
JOIN accounts a ON t.account_id = a.id
WHERE t.date >= ? AND t.date <= ?
GROUP BY a.id
ORDER BY income DESC;

-- name: ReportByTag :many
SELECT
    tg.name as tag,
    SUM(t.amount) as total,
    COUNT(*) as count
FROM transactions t
JOIN transaction_tags tt ON t.id = tt.transaction_id
JOIN tags tg ON tt.tag_id = tg.id
WHERE t.type = 'expense'
  AND t.date >= ? AND t.date <= ?
GROUP BY tg.id
ORDER BY total DESC;

-- name: ReportUntagged :one
SELECT
    'untagged' as tag,
    SUM(t.amount) as total,
    COUNT(*) as count
FROM transactions t
WHERE t.type = 'expense'
  AND t.date >= ? AND t.date <= ?
  AND NOT EXISTS (SELECT 1 FROM transaction_tags tt WHERE tt.transaction_id = t.id);

-- name: ReportTransfers :one
SELECT COALESCE(SUM(amount), 0) as total
FROM transactions
WHERE type = 'transfer'
  AND date >= ? AND date <= ?;
```

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| No data | `No transactions found for the specified period.` | 0 (warn) |
| Invalid month | `Invalid month format. Use YYYY-MM or month name.` | 1 |
| Export failed | `Failed to export: <error>` | 1 |

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Include sub-category rollup in --by category? | → TBD |
| OQ2 | Support --year for annual summary? | → TBD |
