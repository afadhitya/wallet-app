## 1. Query and Model Support

- [x] 1.1 Add sqlc queries to list active unpaused planned payments for a horizon and to list non-archived accounts needed by forecasts.
- [x] 1.2 Regenerate sqlc output and verify generated query types compile.
- [x] 1.3 Add forecast result structs for monthly balance rows, planned payment occurrences, category breakdowns, bill rows, and warnings.

## 2. Forecast Service

- [x] 2.1 Implement service validation for positive `months` values and account resolution for optional account-scoped forecasts.
- [x] 2.2 Implement planned-payment occurrence expansion across the forecast horizon, including recurring payments and month-end clamping behavior.
- [x] 2.3 Implement balance forecast calculation using current account balances, monthly projected income, monthly projected expenses, net movement, ending balances, and warning generation for negative balances.
- [x] 2.4 Implement bill forecast calculation for active unpaused expense planned payments, ordered by due date with running totals.
- [ ] 2.5 Add service tests for default horizon, multi-month recurrence expansion, account filtering, invalid months, no planned payments, and negative-balance warnings.

## 3. CLI Commands and Output

- [ ] 3.1 Replace the placeholder forecast command with `wallet forecast` flags for `--months`/`-n` and `--account`/`-a`.
- [ ] 3.2 Add `wallet forecast bills` with `--months`/`-n` defaulting to 2.
- [ ] 3.3 Implement text output for balance forecasts with projected income, expenses, net movement, ending balances, planned payment details, category breakdowns, empty states, and warnings.
- [ ] 3.4 Implement text output for bills forecasts with date, bill name, amount, running total, total amount, empty states, and planned-payment-only context.
- [ ] 3.5 Implement JSON output for balance and bills forecasts using the existing global `--json` flag.

## 4. Verification

- [ ] 4.1 Add CLI tests for `wallet forecast`, `wallet forecast --months`, `wallet forecast --account`, `wallet forecast bills`, invalid months, JSON output, and empty states.
- [ ] 4.2 Run the full Go test suite.
- [ ] 4.3 Run the project linter and fix reported issues.
- [ ] 4.4 Verify test coverage passes project expectations; if specific code is impractical to test directly, document the exclusion rationale.
- [ ] 4.5 Run OpenSpec validation for the `add-forecasting` change.
