## 1. Query Layer

- [x] 1.1 Add `internal/query/reports.sql` with aggregate queries for income by category, expense by category with parent context, account breakdowns, tag breakdowns, untagged expenses, transfers, transaction counts, and export rows.
- [x] 1.2 Ensure report queries filter non-archived transactions by inclusive date range and optional account ID.
- [x] 1.3 Ensure report aggregates use locked `base_amount` when present and original `amount` otherwise.
- [x] 1.4 Regenerate sqlc output in `internal/gen` and confirm generated report query methods compile.

## 2. Service Layer

- [x] 2.1 Add report service parameter/result types for filters, period resolution, monthly summaries, breakdown rows, account rows, tag rows, and CSV export rows.
- [x] 2.2 Implement month-name, `YYYY-MM`, and explicit `--from`/`--to` period resolution with validation errors for invalid input.
- [x] 2.3 Implement optional account resolution and apply it consistently to summaries, breakdowns, JSON output, and CSV export.
- [x] 2.4 Implement monthly summary calculation for income, expenses, net, transfers, transaction count, and income/expense category totals.
- [x] 2.5 Implement category breakdown with parent-category context, percentages against total expenses, transaction counts, and deterministic ordering.
- [x] 2.6 Implement account breakdown with income, expenses, net, total row data, and foreign-currency context where available.
- [x] 2.7 Implement tag breakdown with tagged rows, `(untagged)` row support, percentages against total expenses, transaction counts, and deterministic ordering.
- [x] 2.8 Implement CSV export row assembly with deterministic tag ordering and clear write/export errors.
- [x] 2.9 Return consistent validation, no-data, unsupported-export, and export-failure errors for CLI rendering.

## 3. CLI Commands

- [x] 3.1 Add and register `wallet report` command with `--month/-m`, `--from`, `--to`, `--account/-a`, `--by`, `--json`, `--export`, and `--output` flags.
- [x] 3.2 Render stable text output for monthly summaries, including income, expenses, net, transfers, transaction count, and no-data warning behavior.
- [x] 3.3 Render stable text output for category, account, and tag breakdowns with amount, percentage or net fields, counts where applicable, and total rows where applicable.
- [x] 3.4 Render report success and error output through the shared AI-native JSON envelope when `--json` is supplied.
- [x] 3.5 Implement `--export csv` command flow with deterministic default file naming and `--output` override.
- [x] 3.6 Reject unsupported `--by` and `--export` values with clear messages and non-zero exit status.

## 4. Tests

- [x] 4.1 Add report service tests for period resolution using month names, `YYYY-MM`, explicit date ranges, invalid month input, and date-range precedence.
- [x] 4.2 Add report service tests for account filtering across summary, breakdown, JSON data structures, and export rows.
- [x] 4.3 Add report service tests for monthly summary totals, income/expense category summaries, transfer exclusion from net, transaction counts, and no-data behavior.
- [x] 4.4 Add report service tests for category breakdown ordering, parent context, percentages, and transaction counts.
- [x] 4.5 Add report service tests for account breakdown income, expenses, net, total rows, and locked base-amount aggregation for foreign-currency transactions.
- [x] 4.6 Add report service tests for tag breakdown ordering, percentages, tagged expenses, and `(untagged)` expenses.
- [x] 4.7 Add report service or focused unit tests for CSV header, row fields, deterministic tag serialization, default filename generation, custom output path behavior, unsupported format errors, and practical export failure handling.
- [x] 4.8 Add CLI integration tests for report summary and breakdown text output, JSON success output, JSON validation errors, validation failures, CSV file creation, and no-data warning behavior.
- [x] 4.9 If any rendering or filesystem failure path is impractical to test reliably, document the reason and exclude only that path from coverage rather than weakening covered service behavior.

## 5. Verification

- [x] 5.1 Run sqlc generation verification if available, or regenerate and inspect generated diffs.
- [x] 5.2 Run `go test ./...` and fix failures.
- [x] 5.3 Run the repository lint command and fix report-related lint failures.
- [x] 5.4 Run the repository coverage command and restore the required coverage gate, documenting any justified exclusions for genuinely hard-to-test paths.
- [x] 5.5 Run `openspec status --change "add-reports-export"` and confirm the change is apply-ready.
