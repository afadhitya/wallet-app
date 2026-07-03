# 05 — Planned Payments

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md)
> Status: 🔴 pending review | Unblocks: 06-forecasting

---

## Objective

Implement planned payments — recurring bills, subscriptions, and one-time future expenses. Track what's due, pay/skip, and manage upcoming obligations.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| P1 | Fulfillment | Manual — `wallet bill pay <id>` | User controls when tx is created |
| P2 | One-time after pay | Auto-archive | Clean, no manual cleanup needed |
| P3 | Skip | `wallet bill skip <id>` — advance next_due once | Granular, skip one occurrence |

---

## Commands

### `wallet bill add`

Create a planned payment.

```
$ wallet bill add "Netflix" 149000 --monthly --day 15 -c subscriptions -a bca
✓ Bill created: Netflix — Rp149.000/month, due 15th
  Next due: 2026-07-15
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--amount` | | Yes (or positional) | — | Amount |
| `--category` | `-c` | No | uncategorized | Category |
| `--account` | `-a` | No | config default | Account |
| `--monthly` | | No | — | Recurrence: monthly |
| `--weekly` | | No | — | Recurrence: weekly |
| `--yearly` | | No | — | Recurrence: yearly |
| `--daily` | | No | — | Recurrence: daily |
| `--rrule` | | No | — | Custom RRULE (RFC 5545) |
| `--day` | | No | today | Day of month/week for recurrence |
| `--from` | | No | today | Start date |
| `--json` | | No | false | JSON output |

**One-time bill:**
```
$ wallet bill add "Flight to Tokyo" 3000000 -c travel -a bca --from 2026-08-15
✓ One-time bill: Flight to Tokyo — Rp3.000.000, due 2026-08-15
```

**Validation:**
- Amount must be > 0
- At least one recurrence flag OR `--from` (one-time)
- Category/account must exist

---

### `wallet bill list`

List planned payments.

```
$ wallet bill list
┌────┬──────────────────┬────────────┬──────────┬────────────┬──────────┐
│ ID │ Name             │ Amount     │ Recur    │ Next Due   │ Status   │
├────┼──────────────────┼────────────┼──────────┼────────────┼──────────┤
│  1 │ Netflix          │ Rp149.000  │ Monthly  │ 2026-07-15 │ Active   │
│  2 │ Gym Membership   │ Rp500.000  │ Monthly  │ 2026-07-01 │ Active   │
│  3 │ Flight to Tokyo  │ Rp3.000.000│ Once     │ 2026-08-15 │ Active   │
│  4 │ Old Subscription │ Rp99.000   │ Monthly  │ 2026-06-15 │ Paused   │
└────┴──────────────────┴────────────┴──────────┴────────────┴──────────┘
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--active` | Show only active, non-paused (default) |
| `--paused` | Show only paused bills |
| `--active --paused` | Show active AND paused (not archived) |
| `--all` | Include paused and archived |
| `--json` | JSON output |

---

### `wallet bill due`

Show upcoming due payments.

```
$ wallet bill due --this-week
┌────┬──────────────────┬────────────┬────────────┬──────────┐
│ ID │ Name             │ Amount     │ Due Date   │ Overdue  │
├────┼──────────────────┼────────────┼────────────┼──────────┤
│  2 │ Gym Membership   │ Rp500.000  │ 2026-07-01 │ 1 day    │
└────┴──────────────────┴────────────┴────────────┴──────────┘
  Total due this week: Rp500.000
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--this-week` | | Due within current week |
| `--this-month` | | Due within current month |
| `--overdue` | | Past due, not yet paid |
| `--next` | `-n` | Next N days |
| `--json` | | JSON output |

**Default:** Show all due this month.

---

### `wallet bill pay <id>`

Fulfill a planned payment — creates a transaction and advances next_due.

```
$ wallet bill pay 1
✓ Paid: Netflix — Rp149.000 (transaction #56 created)
  Next due: 2026-08-15
```

**Behavior:**
1. Load planned payment
2. Create transaction with `is_planned=1`, `planned_payment_id=<id>`
3. If recurring: advance `next_due_date` to next occurrence
4. If one-time: set `is_active=0` (auto-archive)
5. Update account balance

**Flags:**

| Flag | Description |
|------|-------------|
| `--date` | Override transaction date (default: today) |
| `--amount` | Override amount (partial payment) |
| `--json` | JSON output |

---

### `wallet bill skip <id>`

Skip one occurrence — advance next_due without creating transaction.

```
$ wallet bill skip 2
✓ Skipped: Gym Membership — next due advanced to 2026-08-01
```

**Behavior:**
1. Load planned payment
2. Advance `next_due_date` to next occurrence
3. No transaction created

---

### `wallet bill pause <id>`

Pause a recurring bill — stops showing in due lists.

```
$ wallet bill pause 1
✓ Paused: Netflix
```

---

### `wallet bill resume <id>`

Resume a paused bill.

```
$ wallet bill resume 1
✓ Resumed: Netflix — next due: 2026-08-15
```

---

### `wallet bill edit <id>`

Edit a planned payment.

```
$ wallet bill edit 1 --amount 169000
✓ Updated: Netflix — Rp169.000/month
```

**Editable fields:**
- `--amount`
- `--name`
- `--category`
- `--account`
- `--day` (change due day)

---

### `wallet bill rm <id>`

Delete a planned payment.

```
$ wallet bill rm 1
✓ Deleted: Netflix
```

---

## Service Layer

### PlannedPaymentService

```go
type PlannedPaymentService struct {
    db *gen.Queries
}

func (s *PlannedPaymentService) Add(ctx, params) (*PlannedPayment, error)
func (s *PlannedPaymentService) List(ctx, activeOnly) ([]PlannedPayment, error)
func (s *PlannedPaymentService) Due(ctx, filters) ([]PlannedPayment, error)
func (s *PlannedPaymentService) Pay(ctx, id, overrides) (*Transaction, error)
func (s *PlannedPaymentService) Skip(ctx, id) error
func (s *PlannedPaymentService) Pause(ctx, id) error
func (s *PlannedPaymentService) Resume(ctx, id) error
func (s *PlannedPaymentService) Edit(ctx, id, params) (*PlannedPayment, error)
func (s *PlannedPaymentService) Delete(ctx, id) error
func (s *PlannedPaymentService) AdvanceNextDue(ctx, id) error
func (s *PlannedPaymentService) CalcNextDue(ctx, currentDue, recurrence, rule) (time.Time, error)
```

---

## sqlc Queries

### planned_payments.sql

```sql
-- name: GetPlannedPayment :one
SELECT * FROM planned_payments WHERE id = ?;

-- name: ListPlannedPayments :many
SELECT * FROM planned_payments WHERE is_paused = 0 ORDER BY next_due_date;

-- name: ListAllPlannedPayments :many
SELECT * FROM planned_payments ORDER BY next_due_date;

-- name: ListDuePayments :many
SELECT * FROM planned_payments
WHERE is_paused = 0
  AND next_due_date <= ?
ORDER BY next_due_date;

-- name: ListOverduePayments :many
SELECT * FROM planned_payments
WHERE is_paused = 0
  AND next_due_date < date('now')
  AND is_active = 1
ORDER BY next_due_date;

-- name: CreatePlannedPayment :one
INSERT INTO planned_payments (account_id, category_id, type, amount, currency, name, recurrence, recurrence_rule, start_date, next_due_date, is_paused)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0) RETURNING *;

-- name: UpdatePlannedPayment :exec
UPDATE planned_payments SET name = ?, amount = ?, category_id = ?, next_due_date = ?, updated_at = datetime('now') WHERE id = ?;

-- name: UpdateNextDue :exec
UPDATE planned_payments SET next_due_date = ?, updated_at = datetime('now') WHERE id = ?;

-- name: ArchivePlannedPayment :exec
UPDATE planned_payments SET is_active = 0, updated_at = datetime('now') WHERE id = ?;

-- name: PausePlannedPayment :exec
UPDATE planned_payments SET is_paused = 1, updated_at = datetime('now') WHERE id = ?;

-- name: ResumePlannedPayment :exec
UPDATE planned_payments SET is_paused = 0, updated_at = datetime('now') WHERE id = ?;

-- name: DeletePlannedPayment :exec
DELETE FROM planned_payments WHERE id = ?;
```

---

## Next Due Calculation

```
CalcNextDue(currentDue, recurrence, rule):
  switch recurrence:
    daily:    return currentDue + 1 day
    weekly:   return currentDue + 1 week
    monthly:  return currentDue + 1 month (same day, clamp to last day)
    yearly:   return currentDue + 1 year
    custom:   parse RRULE, compute next occurrence
    none:     return nil (one-time, no next due)
```

**Edge cases:**
- Jan 31 + monthly → Feb 28 (clamp to last day of month)
- Feb 28 + monthly → Mar 28 (NOT Mar 31 — keep same day)
- Dec 31 + yearly → next Dec 31

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| Bill not found | `Bill #99 not found.` | 1 |
| Already paid | `Bill #3 already paid for this period. Use --force to re-pay.` | 1 |
| Paused bill | `Bill #1 is paused. Resume it first.` | 1 |
| One-time skip | `Cannot skip one-time bill. Use 'wallet bill rm' to delete.` | 1 |
| Invalid RRULE | `Invalid recurrence rule: <error>` | 1 |

---

## Resolved Questions

| # | Question | Resolution |
|---|----------|------------|
| OQ1 | Allow re-pay same period? | TBD — use `--force` flag for corrections |
| OQ2 | Bill tags support? | TBD — not in current schema, can add later if needed |
