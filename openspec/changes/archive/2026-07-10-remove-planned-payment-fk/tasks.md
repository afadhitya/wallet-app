## 1. Database Migration

- [x] 1.1 Create migration `internal/db/migrations/004_remove_planned_payment_fk.sql` with `ALTER TABLE transactions DROP COLUMN planned_payment_id;` and `ALTER TABLE transactions DROP COLUMN is_planned;`

## 2. SQL Query Changes

- [x] 2.1 Remove the `CreatePlannedTransaction` query from `internal/query/transactions.sql` (lines 4-6)

## 3. Generated Code

- [x] 3.1 Run `sqlc generate` to regenerate `internal/gen/` — verifies that the `Transaction` struct loses `IsPlanned` and `PlannedPaymentID` fields, `CreatePlannedTransaction` is removed from the interface, and all `Scan()` calls adjust to reduced column count

## 4. Service Layer

- [x] 4.1 Update `PayPlannedPayment` in `internal/service/planned_payment.go` to use `CreateTransaction` instead of `CreatePlannedTransaction`, removing the `sql.NullInt64` variable and `CreatePlannedTransactionParams` construction

## 5. Tests

- [x] 5.1 Remove assertions on `IsPlanned` and `PlannedPaymentID` in `internal/service/planned_payment_test.go` (lines 326-330)
- [x] 5.2 Run `make test` to verify all tests pass
- [x] 5.3 Run `make coverage-check` to verify 100% coverage is maintained, if the test is really hard to test, thats fine to exclude that
