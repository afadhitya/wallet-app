# Design: Remove `planned_payment_id` FK from Transactions

## Problem

`bill rm` (hard `DELETE FROM planned_payments`) fails with a SQLite foreign key constraint error when the bill has been paid, because `transactions.planned_payment_id` references `planned_payments(id)` with no `ON DELETE` action.

The FK exists for potential "audit and future reporting" but is never read — no report, budget, statistics, or CLI query joins transactions back to planned payments via this column. It is write-only dead weight that creates a constraint problem.

Rather than work around the FK with soft-delete, remove the useless link entirely.

## Solution

Remove the `planned_payment_id` and `is_planned` columns from the `transactions` table. Without the FK, `bill rm` works freely on paid bills via its existing hard delete.

## Changes

### 1. Migration

New migration `004_remove_planned_payment_fk.sql`:
- `ALTER TABLE transactions DROP COLUMN planned_payment_id;`
- `ALTER TABLE transactions DROP COLUMN is_planned;`

Note: SQLite supports `DROP COLUMN` since 3.35.0 (2021-03-12). The project uses `modernc.org/sqlite` (pure Go), which supports this.

### 2. Query SQL

**`internal/query/transactions.sql`:**
- Remove `CreatePlannedTransaction` query (lines 4-6).
- Remove `is_planned` and `planned_payment_id` from all column lists — currently they are only explicitly listed in `CreatePlannedTransaction`; all other queries use `SELECT *` / `RETURNING *` which will pick up the new schema automatically after regeneration.

### 3. Generated Code (sqlc)

After running `sqlc generate`:
- `gen/models.go` — `Transaction` struct loses `IsPlanned` and `PlannedPaymentID` fields.
- `gen/transactions.sql.go` — `CreatePlannedTransaction` and its params struct are removed. All other `Scan()` calls adjust to the reduced column count.
- `gen/querier.go` — `CreatePlannedTransaction` interface method removed.

### 4. Service Layer

**`internal/service/planned_payment.go`:**
- `PayPlannedPayment()` (line 285-298): Replace `s.q.CreatePlannedTransaction(...)` with `s.q.CreateTransaction(...)`. Remove the `plannedPaymentID` sql.NullInt64 variable and `CreatePlannedTransactionParams` construction. The transaction is created as a plain expense/income/transfer with no planned payment linkage.

**`internal/service/planned_payment_test.go`:**
- Remove assertions on `IsPlanned` and `PlannedPaymentID` (lines 326-330).
- The existing test flow still validates the transaction amount, type, category, account — just without the planned payment link.

### 5. Files NOT Changed

- **`internal/service/forecast.go`** — `PlannedPaymentOccurrence.PlannedPaymentID` is populated from the `planned_payments` table PK, not from `transactions`, so it is unaffected.
- **All report/budget SQL** — never referenced `is_planned` or `planned_payment_id`.
- **CLI layer** — `runBillPay()` accesses `result.Transaction.ID` and `result.Transaction.Amount` only.
- **`internal/service/transaction.go`** — `AddExpense`, `AddIncome`, etc. already use `CreateTransaction` and are unaffected.

## Risk

- **No rollback**: `DROP COLUMN` is irreversible in SQLite. Migrations are forward-only, which is consistent with the existing migration strategy.
- **No data preservation**: The `planned_payment_id` link on past transactions is lost. Since nothing reads it, this is acceptable.
- **No FK safety net**: After this change, deleting a planned payment does not protect linked transactions — but that's the point: transactions are independent and should survive bill deletion.

## Verification

1. `make test` — all existing tests pass with updated assertions
2. `make coverage-check` — 100% coverage maintained
3. Manual: `wallet bill pay <id>` followed by `wallet bill rm <id>` succeeds without FK error
