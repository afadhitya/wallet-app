## Why

`bill rm` fails with a SQLite foreign key constraint error when the bill has been paid, because `transactions.planned_payment_id` references `planned_payments(id)` with no `ON DELETE` action. The FK is write-only dead weight — no report, budget, statistics, or CLI query reads it. Removing the useless link eliminates the constraint problem and simplifies the data model.

## What Changes

- **BREAKING**: Remove `planned_payment_id` and `is_planned` columns from the `transactions` table via a new migration
- Remove the `CreatePlannedTransaction` sqlc query
- Update `PayPlannedPayment` to use the standard `CreateTransaction` instead of `CreatePlannedTransaction`
- Transactions created by `bill pay` are plain expense/income/transfer records with no planned payment linkage

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `wallet-data-model`: Remove the "Planned payment transaction stores source linkage" scenario — transactions no longer reference a source planned payment. Remove the "Archived planned payment remains linkable" scenario — archived planned payments no longer need to be reachable from transactions.
- `planned-payments`: Update the "Pay recurring planned payment" and "Pay one-time planned payment" scenarios — the created transaction is no longer marked as planned or linked to the planned payment.

## Impact

- Database schema: `transactions` table loses two columns
- `internal/query/transactions.sql`: `CreatePlannedTransaction` query removed
- `internal/gen/`: All generated sqlc code regenerated (models, querier, query functions)
- `internal/service/planned_payment.go`: `PayPlannedPayment` simplified
- `internal/service/planned_payment_test.go`: Assertions on `IsPlanned` and `PlannedPaymentID` removed
- All report/budget SQL, forecast logic, and CLI layer are unaffected
