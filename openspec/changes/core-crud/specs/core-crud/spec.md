## ADDED Requirements

### Requirement: Wallet Initialization Command
The system SHALL provide `wallet init` to initialize the local wallet database, apply migrations, seed default categories, and create a default configuration file when missing.

#### Scenario: Initialize a new wallet
- **WHEN** the user runs `wallet init` with no existing database or configuration
- **THEN** the system creates the database at the configured default data path
- **AND** applies all pending migrations
- **AND** seeds the default category hierarchy
- **AND** creates a default config file at the configured default config path

#### Scenario: Re-run initialization safely
- **WHEN** the user runs `wallet init` after the wallet database already exists
- **THEN** the system leaves existing wallet data intact
- **AND** applies only pending migrations
- **AND** does not duplicate seeded categories

### Requirement: Expense And Income Entry Commands
The system SHALL provide `wallet add expense` and `wallet add income` commands that record positive-amount transactions with category, account, tag, date, description, notes, and JSON-output support.

#### Scenario: Add an expense transaction
- **WHEN** the user runs `wallet add expense 35000 "Lunch at Warung" -c food -a bca -t lunch -t work`
- **THEN** the system records an expense transaction for amount `35000`
- **AND** associates the resolved category, account, and tags
- **AND** decreases the selected account balance by the transaction amount
- **AND** prints a success message or JSON representation according to the output mode

#### Scenario: Add an income transaction
- **WHEN** the user runs `wallet add income 5000000 "Gaji Juli" -c salary -a bca`
- **THEN** the system records an income transaction for amount `5000000`
- **AND** associates the resolved category and account
- **AND** increases the selected account balance by the transaction amount

#### Scenario: Reject invalid transaction input
- **WHEN** the user adds an expense or income with a non-positive amount, missing account, missing category, missing tag, or invalid date
- **THEN** the system exits with a non-zero status
- **AND** prints a clear validation error that identifies the invalid field

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
The system SHALL provide `wallet list` to list non-archived transactions with filters for month, account, category, tag, type, date range, limit, and JSON output.

#### Scenario: List current-month transactions by default
- **WHEN** the user runs `wallet list` without filters
- **THEN** the system lists non-archived transactions from the current month
- **AND** orders results by transaction date descending and ID descending
- **AND** limits the output to the default limit of `20`

#### Scenario: List transactions with filters
- **WHEN** the user runs `wallet list --month july --category food --tag lunch --account bca --type expense`
- **THEN** the system returns only non-archived transactions matching all provided filters
- **AND** includes a total for the listed transactions in text output

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

### Requirement: Service Layer And Query Support
The system SHALL implement service-layer operations and sqlc-backed queries for accounts, transactions, categories, tags, transaction tags, and balance recalculation.

#### Scenario: Services validate and persist CRUD operations
- **WHEN** CLI commands invoke service methods for core CRUD operations
- **THEN** the services validate domain inputs before writing
- **AND** use sqlc-generated queries for database access
- **AND** return typed results or clear errors for the CLI to render

#### Scenario: Balance recalculation ignores archived transactions
- **WHEN** an account balance is recalculated
- **THEN** the calculation includes non-archived income, expense, transfer, and adjustment transactions affecting that account
- **AND** excludes archived transactions

### Requirement: Core CRUD Testing
The system SHALL include unit and integration tests for core CRUD service and CLI behavior with deterministic local databases.

#### Scenario: Service tests run against isolated SQLite databases
- **WHEN** service unit tests run
- **THEN** each test uses a clean SQLite database with migrations and seed data applied
- **AND** verifies persisted records and balance changes directly

#### Scenario: CLI integration tests verify command behavior
- **WHEN** CLI integration tests execute core CRUD commands
- **THEN** they verify exit codes, stable output content, JSON output where supported, and database side effects
