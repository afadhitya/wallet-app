## MODIFIED Requirements

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
