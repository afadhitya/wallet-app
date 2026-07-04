## ADDED Requirements

### Requirement: Report Period Filtering
The system SHALL resolve report periods from either `--month` or an explicit inclusive `--from` and `--to` date range, and SHALL support optional account filtering across report summaries, breakdowns, JSON output, and CSV export.

#### Scenario: Month name selects current-year month
- **WHEN** the user runs `wallet report --month july`
- **THEN** the system reports transactions from July of the current year
- **AND** the rendered period label identifies that month and year

#### Scenario: Year-month selects explicit month
- **WHEN** the user runs `wallet report --month 2026-07`
- **THEN** the system reports transactions from `2026-07-01` through `2026-07-31`

#### Scenario: Date range overrides month
- **WHEN** the user runs `wallet report --month 2026-07 --from 2026-07-10 --to 2026-07-20`
- **THEN** the system reports transactions from `2026-07-10` through `2026-07-20`
- **AND** does not include transactions outside that explicit date range

#### Scenario: Account filter limits report data
- **WHEN** the user runs `wallet report --month 2026-07 --account bca`
- **THEN** the system includes only non-archived transactions for the resolved `bca` account in summary totals, breakdowns, JSON output, and CSV export

#### Scenario: Invalid period input is rejected
- **WHEN** the user runs `wallet report --month not-a-month`
- **THEN** the system exits with a non-zero status
- **AND** reports that the month format is invalid and accepts month names or `YYYY-MM`

### Requirement: Monthly Report Summary
The system SHALL provide `wallet report` to summarize non-archived transactions for the selected period, including income, expenses, net, transfer total, and transaction count.

#### Scenario: Generate monthly summary
- **WHEN** the user runs `wallet report --month 2026-07`
- **THEN** the system renders a monthly report with total income, total expenses, net amount, transfer total, and transaction count
- **AND** income and expense totals exclude transfers and adjustments
- **AND** net equals income minus expenses

#### Scenario: Income and expense category summaries are included
- **WHEN** the selected period contains income and expense transactions with categories
- **THEN** the monthly summary includes income and expense category totals
- **AND** orders category totals by amount descending within each section

#### Scenario: Transfers are reported separately
- **WHEN** the selected period contains transfer transactions
- **THEN** the report includes a transfer total for that period
- **AND** transfer amounts do not contribute to income, expenses, or net

#### Scenario: No report data warns without failing
- **WHEN** the selected period contains no matching transactions
- **THEN** the system prints `No transactions found for the specified period.`
- **AND** exits with status `0`

### Requirement: Category Breakdown Report
The system SHALL provide `wallet report --by category` to report expense totals grouped by category with amount, percentage, and transaction count.

#### Scenario: Breakdown by category
- **WHEN** the user runs `wallet report --month 2026-07 --by category`
- **THEN** the system renders expense totals by category for the selected period
- **AND** each row includes category name, amount, percentage of total expenses, and transaction count
- **AND** rows are ordered by amount descending

#### Scenario: Category breakdown includes parent context
- **WHEN** expense transactions use child categories under parent categories
- **THEN** the category breakdown includes parent category context for those child categories
- **AND** the text output can present child rows under their parent category

### Requirement: Account Breakdown Report
The system SHALL provide `wallet report --by account` to report income, expenses, and net amount grouped by account.

#### Scenario: Breakdown by account
- **WHEN** the user runs `wallet report --month 2026-07 --by account`
- **THEN** the system renders one row per account with matching transactions
- **AND** each row includes account name, income, expenses, and net amount
- **AND** a total row summarizes income, expenses, and net across listed accounts

#### Scenario: Foreign-currency account includes base totals
- **WHEN** an account has foreign-currency transactions with locked base amounts
- **THEN** the account breakdown uses locked base-currency equivalents for income, expense, net, and total calculations
- **AND** preserves account currency context where available in structured results

### Requirement: Tag Breakdown Report
The system SHALL provide `wallet report --by tag` to report expense totals grouped by tag, including untagged expense transactions.

#### Scenario: Breakdown by tag
- **WHEN** the user runs `wallet report --month 2026-07 --by tag`
- **THEN** the system renders expense totals by tag for the selected period
- **AND** each row includes tag name, amount, percentage of total expenses, and transaction count
- **AND** rows are ordered by amount descending

#### Scenario: Untagged expenses are included
- **WHEN** the selected period contains expense transactions without tags
- **THEN** the tag breakdown includes an `(untagged)` row
- **AND** untagged expenses contribute to the total used for percentages

### Requirement: Report JSON Output
The system SHALL render report command results through the shared AI-native JSON envelope when `--json` is supplied.

#### Scenario: Monthly report returns JSON envelope
- **WHEN** the user runs `wallet report --month 2026-07 --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains period, income total and breakdown, expense total and breakdown, net, transfers, and transaction count
- **AND** `meta.command` identifies the report command

#### Scenario: Breakdown report returns JSON envelope
- **WHEN** the user runs `wallet report --month 2026-07 --by category --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` identifies the selected breakdown type and contains stable row fields for name, amount, percent, and transaction count

#### Scenario: Report errors return JSON envelope
- **WHEN** the user runs a report command with invalid input and `--json`
- **THEN** the system writes a JSON envelope with `success: false`
- **AND** the error payload identifies the validation failure

### Requirement: CSV Report Export
The system SHALL provide `wallet report --export csv` to export matching report transaction rows to a CSV file, with optional `--output` for custom file paths.

#### Scenario: Export report to default CSV file
- **WHEN** the user runs `wallet report --month 2026-07 --by category --export csv`
- **THEN** the system writes a CSV file using a deterministic default name for the selected period
- **AND** prints a success message identifying the exported path

#### Scenario: Export report to custom CSV file
- **WHEN** the user runs `wallet report --month 2026-07 --export csv --output july-report.csv`
- **THEN** the system writes the CSV to `july-report.csv`
- **AND** reports that output path in the success message

#### Scenario: CSV contains report transaction fields
- **WHEN** the system exports matching report rows to CSV
- **THEN** the first row is `date,type,amount,currency,base_amount,category,account,description,tags`
- **AND** each subsequent row contains one matching transaction with tags serialized in a stable comma-separated order

#### Scenario: Unsupported export format is rejected
- **WHEN** the user runs `wallet report --export pdf`
- **THEN** the system exits with a non-zero status
- **AND** reports that only CSV export is supported

#### Scenario: Export failures are reported
- **WHEN** the CSV file cannot be written
- **THEN** the system exits with a non-zero status
- **AND** reports `Failed to export: <error>`

### Requirement: Report Query And Service Support
The system SHALL implement sqlc queries and service-layer operations for period resolution, account resolution, summaries, category breakdowns, account breakdowns, tag breakdowns, transfers, and CSV export rows.

#### Scenario: Services validate report filters
- **WHEN** CLI commands invoke report service methods
- **THEN** the service validates period, breakdown type, export format, output path, and account filter inputs before executing report queries
- **AND** returns typed results or clear errors for the CLI to render

#### Scenario: Generated queries are current
- **WHEN** report sqlc query files are added or changed
- **THEN** generated code in `internal/gen` is regenerated and compiles with the report service implementation

#### Scenario: Report totals use locked base amounts
- **WHEN** a transaction has a locked `base_amount`
- **THEN** report service totals and percentages use that base amount for aggregation
- **AND** transactions without `base_amount` use their original amount

### Requirement: Report Testing And Quality
The system SHALL include focused tests for report service and CLI behavior and SHALL continue to satisfy repository lint and Go unit coverage requirements, with any genuinely impractical coverage exclusions documented.

#### Scenario: Service tests cover report behavior
- **WHEN** report service tests run
- **THEN** they verify period filtering, account filtering, summary totals, transfers, category breakdowns, account breakdowns, tag breakdowns, untagged rows, mixed-currency base totals, CSV rows, validation errors, and no-data behavior against isolated migrated SQLite databases

#### Scenario: CLI tests cover report commands
- **WHEN** CLI integration tests run report commands
- **THEN** they verify exit codes, stable text output, JSON output, validation errors, CSV file creation, export failure handling where practical, and database read behavior

#### Scenario: Linter passes after report implementation
- **WHEN** the repository lint command runs after report implementation
- **THEN** it completes successfully without report-related lint failures

#### Scenario: Coverage gate passes after report implementation
- **WHEN** the repository coverage command runs after report implementation
- **THEN** total included Go test coverage remains exactly `100%`
- **AND** report service, CLI, database, output, validation, CSV, and error paths are either covered by focused tests or documented as excluded because they are impractical to exercise reliably
