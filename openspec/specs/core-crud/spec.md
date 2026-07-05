## Purpose

TBD

## Requirements

### Requirement: Wallet Initialization Command
The system SHALL provide `wallet init` to initialize the local wallet database, apply migrations, seed default categories, and create default configuration files when missing.

#### Scenario: Initialize a new wallet
- **WHEN** the user runs `wallet init` with no existing database or configuration
- **THEN** the system creates the database at the configured default data path
- **AND** applies all pending migrations
- **AND** seeds the default category hierarchy
- **AND** creates a default config file at the configured default config path
- **AND** creates default rate configuration with base currency `IDR` when rate configuration is missing

#### Scenario: Re-run initialization safely
- **WHEN** the user runs `wallet init` after the wallet database already exists
- **THEN** the system leaves existing wallet data intact
- **AND** applies only pending migrations
- **AND** does not duplicate seeded categories
- **AND** does not overwrite existing rate configuration

### Requirement: Expense And Income Entry Commands
The system SHALL provide `wallet add expense` and `wallet add income` commands that record positive-amount transactions with category, account, tag, date, description, notes, JSON-output support, and locked base-currency conversion for non-base-currency accounts.

#### Scenario: Add an expense transaction
- **WHEN** the user runs `wallet add expense 35000 "Lunch at Warung" -c food -a bca -t lunch -t work`
- **THEN** the system records an expense transaction for amount `35000`
- **AND** associates the resolved category, account, and tags
- **AND** decreases the selected account balance by the transaction amount
- **AND** stores base-currency conversion fields only when the selected account currency differs from the configured base currency
- **AND** prints a success message or JSON representation according to the output mode

#### Scenario: Add an income transaction
- **WHEN** the user runs `wallet add income 5000000 "Gaji Juli" -c salary -a bca`
- **THEN** the system records an income transaction for amount `5000000`
- **AND** associates the resolved category and account
- **AND** increases the selected account balance by the transaction amount
- **AND** stores base-currency conversion fields only when the selected account currency differs from the configured base currency

#### Scenario: Reject invalid transaction input
- **WHEN** the user adds an expense or income with a non-positive amount, missing account, missing category, missing tag, or invalid date
- **THEN** the system exits with a non-zero status
- **AND** prints a clear validation error that identifies the invalid field

#### Scenario: Reject foreign-currency transaction without configured rate
- **WHEN** the user adds an expense or income for an account whose currency differs from the configured base currency and no rate exists for that currency
- **THEN** the system exits with a non-zero status
- **AND** does not create the transaction
- **AND** prints a clear error that identifies the missing rate and how to add it

### Requirement: Transfer Entry Command
The system SHALL provide `wallet add transfer` to record money movement between two different accounts using the source account as `account_id` and destination account as `transfer_to_id`.

#### Scenario: Add a transfer transaction
- **WHEN** the user runs `wallet add transfer 100000 --from bca --to gopay`
- **THEN** the system records a transfer transaction for amount `100000`
- **AND** decreases the source account balance by the transfer amount
- **AND** increases the destination account balance by the transfer amount
- **AND** prints a transfer success message or JSON representation according to the output mode

#### Scenario: Warn on insufficient source balance
- **WHEN** the user records a transfer whose amount is greater than the source account balance
- **THEN** the system records the transfer successfully
- **AND** prints a warning about the insufficient source balance
- **AND** exits with status `0`

#### Scenario: Reject invalid transfer accounts
- **WHEN** the user records a transfer with missing accounts or the same source and destination account
- **THEN** the system exits with a non-zero status
- **AND** does not create a transaction

### Requirement: Transaction Listing Command
The system SHALL provide `wallet list` to list non-archived transactions with filters for month, account, category, tag, type, date range, limit, JSON output, and locked base-currency equivalents for converted transactions.

#### Scenario: List current-month transactions by default
- **WHEN** the user runs `wallet list` without filters
- **THEN** the system lists non-archived transactions from the current month
- **AND** orders results by transaction date descending and ID descending
- **AND** limits the output to the default limit of `20`
- **AND** displays locked base-currency equivalents for converted transactions where present

#### Scenario: List transactions with filters
- **WHEN** the user runs `wallet list --month july --category food --tag lunch --account bca --type expense`
- **THEN** the system returns only non-archived transactions matching all provided filters
- **AND** includes a total for the listed transactions in text output
- **AND** includes base-currency totals when listed transactions include converted amounts

### Requirement: Transaction Editing Command
The system SHALL provide `wallet edit <id>` to update only explicitly supplied transaction fields and update affected account balances.

#### Scenario: Edit transaction amount and category
- **WHEN** the user runs `wallet edit 42 --amount 40000 --category transport`
- **THEN** the system updates transaction `42` with the new amount and category
- **AND** updates the transaction `updated_at` timestamp
- **AND** recalculates balances for every affected account

#### Scenario: Edit transaction tags
- **WHEN** the user runs `wallet edit 42 --add-tag work --remove-tag lunch`
- **THEN** the system adds the resolved `work` tag to transaction `42`
- **AND** removes the resolved `lunch` tag from transaction `42`
- **AND** leaves unspecified transaction fields unchanged

#### Scenario: Reject editing a missing transaction
- **WHEN** the user runs `wallet edit 99 --amount 40000` and transaction `99` does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that transaction `99` was not found

### Requirement: Transaction Removal Command
The system SHALL provide `wallet rm <id>` to delete transactions by marking them archived and recalculating affected account balances.

#### Scenario: Remove transaction with confirmation bypass
- **WHEN** the user runs `wallet rm 42 --force`
- **THEN** the system marks transaction `42` as archived
- **AND** recalculates balances for every affected account
- **AND** excludes transaction `42` from default list results

#### Scenario: Confirm transaction removal interactively
- **WHEN** the user runs `wallet rm 42` without `--force`
- **THEN** the system prompts for confirmation before archiving the transaction
- **AND** does not archive the transaction if the user declines

### Requirement: Category Management Commands
The system SHALL provide `wallet category list`, `wallet category add`, `wallet category edit <id>`, and category removal behavior for managing two-level categories.

#### Scenario: List categories
- **WHEN** the user runs `wallet category list`
- **THEN** the system lists active parent and child categories with ID, name, parent, type, and icon fields

#### Scenario: Add category
- **WHEN** the user runs `wallet category add "Kopi" --parent "Food & Dining" --icon "coffee"`
- **THEN** the system creates a child category named `Kopi` under `Food & Dining`
- **AND** makes it available for transaction entry

#### Scenario: Edit category
- **WHEN** the user runs `wallet category edit 3 --name "Groceries & Supermarket" --icon "cart"`
- **THEN** the system updates category `3` with only the supplied fields
- **AND** preserves existing transaction links to category `3`

#### Scenario: Remove category
- **WHEN** the user removes a category that is not needed for active entry
- **THEN** the system prevents new transactions from using that category
- **AND** preserves historical transactions that already reference that category

### Requirement: Tag Management Commands
The system SHALL provide `wallet tag list`, `wallet tag add`, and `wallet tag rm <name>` for explicit tag management.

#### Scenario: Add tag
- **WHEN** the user runs `wallet tag add "japan-2026"`
- **THEN** the system creates a unique tag named `japan-2026`
- **AND** makes it available for transaction entry and filtering

#### Scenario: List tags
- **WHEN** the user runs `wallet tag list`
- **THEN** the system lists available tags with ID, name, and color fields

#### Scenario: Remove tag
- **WHEN** the user runs `wallet tag rm "japan-2026"`
- **THEN** the system deletes the tag named `japan-2026`
- **AND** removes its transaction associations through referential cleanup

### Requirement: Balance Adjustment Command
The system SHALL provide `wallet adjust <account> <amount> <description>` to reconcile an account to an exact target balance and record an adjustment transaction.

#### Scenario: Increase account balance through adjustment
- **WHEN** the user runs `wallet adjust bca 15000000 "Found missing deposit"`
- **THEN** the system sets the BCA account balance to `15000000`
- **AND** records an adjustment transaction for the positive difference
- **AND** prints the old balance, new balance, and difference

#### Scenario: Decrease account balance through adjustment
- **WHEN** the user runs `wallet adjust gopay 50000 "Cash out not recorded"`
- **THEN** the system sets the GoPay account balance to `50000`
- **AND** records an adjustment transaction for the absolute difference
- **AND** makes the adjustment visible through `wallet list --type adjustment`

### Requirement: Account Management Commands
The system SHALL provide account lifecycle management through a `wallet account` command group with `add`, `list`, `edit`, and `archive` subcommands.

#### Scenario: Account command group is available
- **WHEN** the user runs `wallet account --help`
- **THEN** the system shows subcommands for add, list, edit, and archive

#### Scenario: Account commands follow existing core CRUD patterns
- **WHEN** the user invokes any account subcommand
- **THEN** the system validates inputs before persisting changes
- **AND** uses the existing account service methods for all operations
- **AND** renders text tables or JSON output matching the existing core CRUD command style
- **AND** returns typed errors (not found, validation, duplicate) matching the existing error classification

#### Scenario: Account commands require initialized wallet
- **WHEN** the user runs an account command before running `wallet init`
- **THEN** the system exits with a non-zero status
- **AND** reports that the database is not initialized

### Requirement: Core CRUD AI-Native JSON Output
The system SHALL render core CRUD command results and failures through the shared AI-native JSON envelope when `--json` is supplied.

#### Scenario: Add transaction returns envelope JSON
- **WHEN** the user runs `wallet add expense 35000 "Lunch at Warung" -c food -a bca --json`
- **THEN** the system records the transaction normally
- **AND** writes a JSON envelope with `success: true`
- **AND** `data` contains the created transaction fields including ID, type, amount, description, category, account, tags, and planned-payment state where applicable
- **AND** `meta.command` identifies the add command

#### Scenario: List transactions returns envelope JSON
- **WHEN** the user runs `wallet list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.transactions` contains the listed transactions
- **AND** `data.total` contains the listed transaction total
- **AND** `data.count` contains the number of listed transactions

#### Scenario: Transfer returns envelope JSON
- **WHEN** the user runs `wallet add transfer 100000 --from bca --to gopay --json`
- **THEN** the system records the transfer normally
- **AND** writes a JSON envelope with `success: true`
- **AND** `data` identifies the transfer transaction, source account, destination account, amount, and any warnings

#### Scenario: Core CRUD errors return envelope JSON
- **WHEN** the user runs a core CRUD command with `--json` and references a missing category, account, tag, or transaction
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the missing or invalid resource type
- **AND** `error.message` describes the failure without table formatting

### Requirement: Service Layer And Query Support
The system SHALL implement service-layer operations and sqlc-backed queries for accounts, transactions, categories, tags, transaction tags, balance recalculation, and transaction-time currency conversion.

#### Scenario: Services validate and persist CRUD operations
- **WHEN** CLI commands invoke service methods for core CRUD operations
- **THEN** the services validate domain inputs before writing
- **AND** use sqlc-generated queries for database access
- **AND** return typed results or clear errors for the CLI to render
- **AND** apply currency conversion through the currency service before persisting non-base-currency income and expense transactions

#### Scenario: Balance recalculation ignores archived transactions
- **WHEN** an account balance is recalculated
- **THEN** the calculation includes non-archived income, expense, transfer, and adjustment transactions affecting that account
- **AND** excludes archived transactions
- **AND** uses original account-currency transaction amounts rather than converted base amounts for account balances

### Requirement: Core CRUD Testing
The system SHALL include unit and integration tests for core CRUD service and CLI behavior with deterministic local databases and SHALL satisfy the repository's 100% Go coverage gate after approved generated-code and documented OS/infrastructure exclusions are applied.

#### Scenario: Service tests run against isolated SQLite databases
- **WHEN** service unit tests run
- **THEN** each test uses a clean SQLite database with migrations and seed data applied
- **AND** verifies persisted records and balance changes directly

#### Scenario: CLI integration tests verify command behavior
- **WHEN** CLI integration tests execute core CRUD commands
- **THEN** they verify exit codes, stable output content, JSON output where supported, and database side effects

#### Scenario: Coverage gate passes in CI
- **WHEN** GitHub Actions runs the repository coverage check for the core CRUD implementation
- **THEN** total included Go test coverage is exactly `100%`
- **AND** generated sqlc code and documented OS/infrastructure failure branches are excluded by an auditable coverage policy
- **AND** every uncovered function, branch, error path, command path, and helper outside the approved exclusion list has either targeted test coverage or has been removed as unreachable code

#### Scenario: Coverage gap is found after implementation
- **WHEN** local or CI coverage reports less than `100%`
- **THEN** the implementation is not considered complete
- **AND** the coverage profile is used to add focused tests for the missing service, CLI, database, output, validation, or error-handling paths unless the gap is an approved generated-code or OS/infrastructure exclusion

#### Scenario: Generated query package is excluded from coverage
- **WHEN** the CI coverage command builds its package list
- **THEN** `internal/gen` sqlc-generated files are excluded from coverage totals
- **AND** generated query code is still compiled by normal tests and checked for staleness by generation verification

#### Scenario: Test database helper package is included in coverage
- **WHEN** the CI coverage profile includes `internal/testdb/testdb.go`
- **THEN** package-local tests exercise migrated database setup, cleanup registration, query construction, and failure handling paths that are reachable without corrupting global test state

#### Scenario: CLI branch coverage gaps are reported
- **WHEN** the coverage profile reports zero-count branches in CLI command handlers or CLI helper functions
- **THEN** CLI tests cover JSON and text output, argument parsing failures, invalid ID parsing, service-construction failures, confirmation accept/decline paths, stdin scanner errors where feasible, and command execution errors
- **AND** any remaining CLI init/mkdir/rm/tag OS-level infrastructure failure branches are documented as approved exclusions before being removed from coverage totals

#### Scenario: OS infrastructure branch is excluded
- **WHEN** a branch represents deterministic OS-level failure handling for CLI init, directory creation, removal, or tag infrastructure that cannot be exercised without brittle fault injection
- **THEN** the branch MAY be excluded from coverage totals
- **AND** the exclusion identifies the file/function or coverage block being excluded
- **AND** the exclusion does not apply to business validation, service behavior, JSON/text rendering, or normal command error handling

### Requirement: Planned Payment Service And Query Support
The system SHALL implement service-layer operations and sqlc-backed queries for planned payment creation, listing, due filtering, fulfillment, skipping, pausing, resuming, editing, deleting, and recurrence calculation.

#### Scenario: Services validate and persist planned payment operations
- **WHEN** CLI commands invoke planned-payment service methods
- **THEN** the services validate domain inputs before writing
- **AND** use sqlc-generated queries for database access
- **AND** return typed results or clear errors for the CLI to render

#### Scenario: Pay operation updates transaction and planned payment state
- **WHEN** the planned-payment service pays a bill
- **THEN** it creates the linked transaction
- **AND** updates affected account balance through existing transaction balance behavior
- **AND** advances or archives the planned payment in the same logical operation

#### Scenario: Due filters ignore inactive and paused payments
- **WHEN** the planned-payment service lists due payments
- **THEN** it excludes paused planned payments
- **AND** excludes archived planned payments

### Requirement: Planned Payment CLI Commands
The system SHALL expose planned-payment workflows through a `wallet bill` command group.

#### Scenario: Bill command group is available
- **WHEN** the user runs `wallet bill --help`
- **THEN** the system shows subcommands for add, list, due, pay, skip, pause, resume, edit, and rm

#### Scenario: Bill commands render text and JSON output
- **WHEN** the user runs a bill command that supports `--json`
- **THEN** the system renders JSON output instead of table or prose output

#### Scenario: Bill commands report missing records
- **WHEN** the user runs a bill command with an identifier that does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that the bill was not found

### Requirement: Planned Payment Testing
The system SHALL include unit and integration tests for planned payment service and CLI behavior with deterministic local databases and SHALL satisfy the repository's 100% Go coverage gate after approved generated-code and documented OS/infrastructure exclusions are applied.

#### Scenario: Service tests cover recurrence and state transitions
- **WHEN** planned-payment service tests run
- **THEN** they verify creation validation, due filtering, pay, skip, pause, resume, edit, delete, and recurrence edge cases against an isolated SQLite database

#### Scenario: CLI integration tests cover bill workflows
- **WHEN** CLI integration tests execute planned-payment commands
- **THEN** they verify exit codes, stable output content, JSON output where supported, and database side effects

#### Scenario: Coverage gate passes for planned payment implementation
- **WHEN** GitHub Actions runs the repository coverage check after planned payments are implemented
- **THEN** total included Go test coverage remains exactly `100%`
- **AND** generated sqlc code and documented OS/infrastructure failure branches remain excluded by the approved coverage policy
