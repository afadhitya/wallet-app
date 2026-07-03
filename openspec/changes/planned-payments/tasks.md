## 1. Schema And Queries

- [x] 1.1 Inspect current migrations for existing `planned_payments` columns and transaction planned-linkage fields.
- [x] 1.2 Add a migration for any missing `planned_payments` active-state fields, transaction `is_planned` and `planned_payment_id` linkage, indexes, and referential constraints.
- [x] 1.3 Add `internal/query/planned_payments.sql` with get, list, due, overdue, create, update, update-next-due, archive, pause, resume, and delete queries.
- [x] 1.4 Extend transaction insert/query definitions to persist and return planned-payment linkage fields where required.
- [x] 1.5 Regenerate sqlc output and verify generated models/queries compile.

## 2. Planned Payment Service

- [x] 2.1 Add `PlannedPaymentService` types and constructor wiring alongside existing services.
- [x] 2.2 Implement create validation for amount, account, category, schedule flags, start date, recurrence type, and custom RRULE values.
- [x] 2.3 Implement list and due filtering for default active, paused, all, current week, current month, overdue, and next-N-days views.
- [x] 2.4 Implement next-due calculation for daily, weekly, monthly with end-of-month clamping, yearly, custom RRULE, and one-time schedules.
- [x] 2.5 Implement pay behavior that creates a linked expense transaction, updates account balance through existing transaction behavior, and advances recurring bills or archives one-time bills.
- [x] 2.6 Implement skip behavior that advances one recurring occurrence without creating a transaction and rejects one-time bills.
- [x] 2.7 Implement pause, resume, edit, and delete behavior with clear not-found and invalid-state errors.

## 3. CLI Commands

- [x] 3.1 Add the `wallet bill` command group and register add, list, due, pay, skip, pause, resume, edit, and rm subcommands.
- [x] 3.2 Implement `wallet bill add` flags for amount, category, account, recurrence, RRULE, due day, start date, and JSON output.
- [x] 3.3 Implement `wallet bill list` filters for active, paused, all, and JSON output with stable text table rendering.
- [x] 3.4 Implement `wallet bill due` filters for current week, current month, overdue, next N days, and JSON output with text totals.
- [x] 3.5 Implement `wallet bill pay` with date and amount overrides, JSON output, linked transaction reporting, and paused/not-found errors.
- [x] 3.6 Implement `wallet bill skip`, `pause`, `resume`, `edit`, and `rm` command behavior and output.

## 4. Tests And Verification

- [ ] 4.1 Add migration/query tests for planned-payment storage, active/paused filters, and transaction planned-payment linkage.
- [ ] 4.2 Add service tests for create validation, recurrence edge cases, due filters, pay, skip, pause, resume, edit, delete, and one-time archiving.
- [ ] 4.3 Add CLI integration tests for representative `wallet bill` workflows, JSON output, validation failures, and missing-record errors.
- [ ] 4.4 Run sqlc generation verification, Go tests, and the repository coverage command until the existing 100% coverage gate passes.
- [ ] 4.5 Run OpenSpec validation/status checks for `planned-payments` and resolve any artifact or requirement formatting issues.
