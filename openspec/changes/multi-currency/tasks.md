## 1. Configuration And Currency Service

- [x] 1.1 Add rate configuration model and TOML load/save support for base currency plus configured rates.
- [x] 1.2 Update `wallet init` to create default rate configuration when missing without overwriting existing rate config.
- [x] 1.3 Implement currency service methods for base currency, rate lookup, conversion, listing, add, set, and remove.
- [x] 1.4 Add validation and actionable errors for missing rate config, missing currency rates, and non-positive rates.
- [x] 1.5 Add unit tests for config parsing, config persistence, conversion identity for base currency, conversion using configured rates, and error cases.

## 2. Rate CLI Commands

- [x] 2.1 Add `wallet rate list` text output showing currencies, rates toward the base currency, inverse rates, and base currency.
- [x] 2.2 Add `wallet rate list --json` output with base currency and configured rates.
- [x] 2.3 Add `wallet rate set <currency> <rate>` to update existing configured rates.
- [x] 2.4 Add `wallet rate add <currency> <rate>` to add configured rates.
- [x] 2.5 Add `wallet rate rm <currency>` to remove configured rates without changing existing transactions.
- [x] 2.6 Add CLI tests for success, JSON output, invalid rates, missing rates, and persistence to rate config.

## 3. Transaction Conversion Integration

- [x] 3.1 Wire the currency service into income and expense creation paths.
- [x] 3.2 For non-base-currency accounts, store original `amount` and `currency` plus locked `base_amount` and `base_currency` at transaction creation time.
- [x] 3.3 For base-currency accounts, keep `base_amount` and `base_currency` unset.
- [x] 3.4 Ensure account balance changes and balance recalculation continue to use original account-currency amounts, not converted base amounts.
- [x] 3.5 Reject non-base-currency income and expense creation when the account currency has no configured rate, without persisting a transaction.
- [x] 3.6 Add service and CLI tests for foreign-currency expense, foreign-currency income, base-currency transaction behavior, balance updates, and missing-rate rejection.

## 4. Listing And Reporting Output

- [x] 4.1 Update transaction query models and renderers to expose locked base-currency equivalents where present.
- [x] 4.2 Update `wallet list` text output to show original amounts and base-currency equivalents for converted transactions.
- [x] 4.3 Update list totals to include base-currency totals when converted amounts are present and original-currency totals when a single foreign currency is listed.
- [x] 4.4 Update JSON output to include original currency fields and locked base-currency fields consistently.
- [x] 4.5 Update reporting aggregation to use base amounts for converted transactions and original amounts for base-currency transactions.
- [x] 4.6 Ensure report income/expense totals continue to exclude adjustment transactions.
- [x] 4.7 Add tests for mixed-currency list output, JSON output, reporting totals, original-currency context, and adjustment exclusion.

## 5. Quality Gates

- [ ] 5.1 Run code generation or query generation if data access signatures changed.
- [ ] 5.2 Run formatting and linting checks used by the repository and ensure the linter passes.
- [ ] 5.3 Run the full Go test suite.
- [ ] 5.4 Run the repository coverage gate and add focused tests until the accepted coverage policy passes.
- [ ] 5.5 If code is genuinely hard to test without brittle fault injection or OS-level manipulation, document the specific exclusion and keep it out of coverage totals only when it is not business validation, service behavior, rendering, or normal command error handling.
- [ ] 5.6 Run `openspec status --change "multi-currency"` and confirm the change is apply-ready.
