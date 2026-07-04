## Why

Users can record transactions, categories, accounts, tags, and locked base-currency amounts, but they cannot summarize that activity into financial reports or export report data for external analysis. Adding reports and CSV export completes the MVP flow from data entry to insight and portable output.

## What Changes

- Add `wallet report` for monthly financial summaries over income, expenses, net, transfers, and transaction count.
- Add report filtering by month, explicit date range, and account.
- Add `wallet report --by category`, `--by account`, and `--by tag` breakdowns.
- Add `wallet report --json` using the shared AI-native JSON envelope.
- Add `wallet report --export csv` with optional `--output` for writing report transactions to CSV.
- Add report service logic, sqlc queries, CLI rendering, validation, error handling, and tests.

## Capabilities

### New Capabilities
- `reports-export`: Financial report commands, report aggregation behavior, CSV export behavior, JSON output, validation, and test expectations.

### Modified Capabilities

## Impact

- Affected CLI: new `wallet report` command with filtering, breakdown, JSON, export, and output flags.
- Affected services: new report service that aggregates existing transaction, category, account, tag, transfer, and currency fields.
- Affected database access: new sqlc queries for report totals, breakdowns, transfers, tag rollups, untagged expenses, and export rows.
- Affected output: text tables, AI-native JSON envelope output, and CSV file generation.
- Affected tests: service unit tests, CLI integration/output tests, CSV export tests, linter checks, and coverage gate verification where practical.
