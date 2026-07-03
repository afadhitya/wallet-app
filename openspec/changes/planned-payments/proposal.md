## Why

Wallet users need a way to track upcoming obligations such as subscriptions, recurring bills, and known future expenses before they become transactions. Adding planned payments makes the CLI useful for managing cash commitments, reviewing what is due soon, and converting a due obligation into an actual transaction when the user chooses to pay it.

## What Changes

- Add a planned payment model for recurring and one-time future expenses.
- Add `wallet bill` commands to create, list, inspect due, pay, skip, pause, resume, edit, and delete planned payments.
- Support daily, weekly, monthly, yearly, custom RRULE, and one-time schedules.
- Convert paid planned payments into expense transactions linked back to their source planned payment.
- Advance recurring payments after pay or skip, and automatically archive one-time payments after pay.
- Exclude paused and archived payments from default active and due views.

## Capabilities

### New Capabilities

- `planned-payments`: Tracks scheduled bills and future expenses, exposes bill-management CLI commands, and supports manually fulfilling due payments into transactions.

### Modified Capabilities

- `wallet-data-model`: Adds planned payment persistence and transaction linkage fields required to relate paid transactions to planned payments.
- `core-crud`: Adds planned-payment service behavior and CLI workflows that depend on existing account, category, transaction, and balance update behavior.

## Impact

- Database migrations and sqlc queries for planned payments and planned transaction linkage.
- Service layer for planned payment validation, recurrence calculation, due filtering, pay, skip, pause, resume, edit, and delete operations.
- CLI command tree under `wallet bill` with table and JSON output modes.
- Transaction creation and account balance updates when a planned payment is paid.
- Tests for recurrence edge cases, command behavior, validation failures, and transaction linkage.
