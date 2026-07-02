# 06 — Forecasting

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md), [05-planned-payments](./05-planned-payments.md)
> Status: 🔴 pending review | Unblocks: 07-multi-currency

---

## Objective

Implement financial forecasting — project future balances based on historical spending and planned payments. Help user anticipate cash flow needs.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| F1 | Forecast source | Planned payments only | Simpler, more predictable, user-controlled accuracy |
| F2 | Horizon | Configurable `--months N` (default 1) | Flexible for different planning needs |

---

## Commands

### `wallet forecast`

Project future balance and spending.

```
$ wallet forecast
┌─────────────────────────────────────────────────────────────┐
│ Balance Forecast — Next 1 Month (Aug 2026)                  │
├─────────────────────────────────────────────────────────────┤
│ Starting Balance:     Rp15.000.000                          │
│                                                             │
│ Projected Income:     Rp5.000.000                           │
│   Salary (planned):   Rp5.000.000                           │
│                                                             │
│ Projected Expenses:   Rp2.649.000                           │
│   Netflix:            Rp149.000 (Aug 15)                    │
│   Gym:                Rp500.000 (Aug 05)                    │
│   Rent:               Rp2.000.000 (Aug 01)                  │
│                                                             │
│ Ending Balance:       Rp17.351.000                          │
│                                                             │
│ ⚠️  Bills Due: Rp1.500.000 on Aug 1, Aug 15                │
└─────────────────────────────────────────────────────────────┘
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--months` | `-n` | 1 | Forecast horizon in months |
| `--account` | `-a` | all | Specific account |
| `--json` | | false | JSON output |

---

### `wallet forecast --months 3`

Multi-month forecast with monthly breakdown.

```
$ wallet forecast --months 3
┌───────────────────────────────────────────────────────────────┐
│ Balance Forecast — 3 Months (Aug–Oct 2026)                    │
├───────────┬─────────────┬─────────────┬─────────────┬────────┤
│ Month     │ Income      │ Expenses    │ Net         │ Balance│
├───────────┼─────────────┼─────────────┼─────────────┼────────┤
│ Aug 2026  │ Rp5.000.000 │ Rp4.200.000 │ +Rp800.000  │Rp15.8M │
│ Sep 2026  │ Rp5.000.000 │ Rp4.100.000 │ +Rp900.000  │Rp16.7M │
│ Oct 2026  │ Rp5.000.000 │ Rp4.300.000 │ +Rp700.000  │Rp17.4M │
└───────────┴─────────────┴─────────────┴─────────────┴────────┘
│ Based on: 5 planned payments |
```

---

### `wallet forecast --account bca`

Per-account forecast.

```
$ wallet forecast --account bca --months 2
┌─────────────────────────────────────────────────────────────┐
│ BCA Forecast — 2 Months (Aug–Sep 2026)                      │
├───────────┬─────────────┬─────────────┬─────────────┬───────┤
│ Month     │ Income      │ Expenses    │ Net         │Balance│
├───────────┼─────────────┼─────────────┼─────────────┼───────┤
│ Aug 2026  │ Rp5.000.000 │ Rp3.800.000 │+Rp1.200.000 │Rp12.2M│
│ Sep 2026  │ Rp5.000.000 │ Rp3.600.000 │+Rp1.400.000 │Rp13.6M│
└───────────┴─────────────┴─────────────┴─────────────┴───────┘
```

---

### `wallet forecast bills`

Show upcoming bills impact on forecast.

```
$ wallet forecast bills --months 2
┌──────────────────────────────────────────────────────────────────┐
│ Bills Forecast — 2 Months                                        │
├───────────┬──────────────────┬────────────┬──────────────────────┤
│ Date      │ Bill             │ Amount     │ Running Total        │
├───────────┼──────────────────┼────────────┼──────────────────────┤
│ Aug 01    │ Rent             │ Rp2.000.000│ Rp2.000.000          │
│ Aug 05    │ Gym              │ Rp500.000  │ Rp2.500.000          │
│ Aug 15    │ Netflix          │ Rp149.000  │ Rp2.649.000          │
│ Sep 01    │ Rent             │ Rp2.000.000│ Rp4.649.000          │
│ Sep 05    │ Gym              │ Rp500.000  │ Rp5.149.000          │
│ Sep 15    │ Netflix          │ Rp149.000  │ Rp5.298.000          │
└───────────┴──────────────────┴────────────┴──────────────────────┘
  Total upcoming bills: Rp5.298.000
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--months` | `-n` | 2 | Forecast horizon |
| `--json` | | false | JSON output |

---

## Service Layer

### ForecastService

```go
type ForecastService struct {
    db          *gen.Queries
    billService *PlannedPaymentService
}

type ForecastResult struct {
    Month           string
    AccountID       int64
    AccountName     string
    StartingBalance int64
    ProjectedIncome int64
    ProjectedExpense int64
    PlannedPayments []PlannedPaymentDetail
    EndingBalance   int64
}

type PlannedPaymentDetail struct {
    Name   string
    Amount int64
    Date   string
}

type BillForecast struct {
    Date        string
    BillName    string
    Amount      int64
    RunningTotal int64
}

func (s *ForecastService) Forecast(ctx, accountID or all, months) ([]ForecastResult, error)
func (s *ForecastService) BillsForecast(ctx, months) ([]BillForecast, error)
```

---

## Forecast Logic

### Balance Projection

```
For each month in horizon:
  1. StartingBalance = previous month's EndingBalance (or current balance)
  2. ProjectedIncome = sum of planned income payments for this month
  3. ProjectedExpense = sum of planned expense payments for this month
  4. EndingBalance = StartingBalance + ProjectedIncome - ProjectedExpense
```

**Notes:**
- Forecast accuracy depends on completeness of planned payments
- Unplanned spending (e.g., impulsive purchases) is NOT projected
- User should add all recurring bills as planned payments for accurate forecasts

---

## sqlc Queries

### forecasting.sql

```sql
-- name: SumPlannedByMonth :many
SELECT
    strftime('%Y-%m', next_due_date) as month,
    SUM(amount) as total,
    type
FROM planned_payments
WHERE is_active = 1
  AND is_paused = 0
  AND next_due_date >= date('now')
  AND next_due_date < date('now', ? || ' months')
GROUP BY strftime('%Y-%m', next_due_date), type
ORDER BY month;

-- name: GetCurrentBalance :one
SELECT balance FROM accounts WHERE id = ? AND is_archived = 0;
```

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| No planned payments | `No planned payments found. Add bills with: wallet bill add` | 0 (warn) |
| Account not found | `Account 'foo' not found.` | 1 |
| Negative projection | `Warning: Projected negative balance in <month>` | 0 (warn) |

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Should forecast include category-level breakdown? | → TBD |
| OQ2 | Include tags in forecast (e.g., #japan-2026 spending)? | → TBD |
