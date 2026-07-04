## ADDED Requirements

### Requirement: Rate Configuration
The system SHALL maintain local exchange-rate configuration with a single base currency and manually managed rates.

#### Scenario: Initialize default rate configuration
- **WHEN** the user runs `wallet init` and no rate configuration exists
- **THEN** the system creates rate configuration at the configured wallet config location
- **AND** the default base currency is `IDR`
- **AND** the configuration stores rates as base-currency units per one foreign-currency unit

#### Scenario: Preserve existing rate configuration
- **WHEN** the user runs `wallet init` and rate configuration already exists
- **THEN** the system leaves the existing base currency and rates unchanged

### Requirement: Rate Management Commands
The system SHALL provide `wallet rate list`, `wallet rate set`, `wallet rate add`, and `wallet rate rm` commands for managing configured manual exchange rates.

#### Scenario: List configured rates
- **WHEN** the user runs `wallet rate list`
- **THEN** the system lists configured currencies, each rate toward the base currency, and the inverse base-to-foreign value
- **AND** it displays the configured base currency

#### Scenario: List configured rates as JSON
- **WHEN** the user runs `wallet rate list --json`
- **THEN** the system emits a JSON object containing the base currency and configured rates

#### Scenario: Set an existing rate
- **WHEN** the user runs `wallet rate set USD 15800`
- **THEN** the system updates the configured USD rate to `15800`
- **AND** future conversions use the new rate
- **AND** existing transactions are not modified

#### Scenario: Add a new rate
- **WHEN** the user runs `wallet rate add KRW 12`
- **THEN** the system adds `KRW` to the configured rates with value `12`
- **AND** future KRW transactions can be converted when KRW differs from the base currency

#### Scenario: Remove a rate
- **WHEN** the user runs `wallet rate rm KRW`
- **THEN** the system removes `KRW` from the configured rates
- **AND** existing transactions that already stored KRW base amounts are not modified

#### Scenario: Reject invalid rate values
- **WHEN** the user adds or sets a rate with a non-positive or non-numeric value
- **THEN** the system exits with a non-zero status
- **AND** reports that the rate must be a positive number

### Requirement: Transaction-Time Currency Conversion
The system SHALL convert foreign-currency income and expense transactions to the configured base currency when the transaction is created.

#### Scenario: Create a foreign-currency expense
- **WHEN** the user records an expense against an account whose currency differs from the configured base currency
- **THEN** the system looks up the account currency in the configured rates
- **AND** stores the original amount and account currency on the transaction
- **AND** stores the converted `base_amount` and configured `base_currency` on the transaction
- **AND** decreases the account balance by the original amount in the account currency

#### Scenario: Create a base-currency expense
- **WHEN** the user records an expense against an account whose currency equals the configured base currency
- **THEN** the system stores the original amount and currency on the transaction
- **AND** leaves `base_amount` and `base_currency` unset
- **AND** decreases the account balance by the original amount

#### Scenario: Reject transaction when rate is missing
- **WHEN** the user records an income or expense against a non-base-currency account and no configured rate exists for that account currency
- **THEN** the system exits with a non-zero status
- **AND** does not create the transaction
- **AND** reports an actionable error that tells the user to add the missing rate

### Requirement: Mixed-Currency Display And Reporting
The system SHALL present mixed-currency transaction and report output with base-currency totals while preserving original-currency context.

#### Scenario: List foreign-currency transactions
- **WHEN** the user lists transactions for a foreign-currency account
- **THEN** each converted transaction displays the original amount in the account currency
- **AND** displays the locked base-currency equivalent when present
- **AND** text totals include both original-currency and base-currency totals when all listed transactions are from the same foreign currency

#### Scenario: Report mixed-currency activity
- **WHEN** the user generates a report that includes multiple transaction currencies
- **THEN** income, expense, net, and balance totals are reported primarily in the configured base currency
- **AND** categories or accounts with foreign-currency activity include original-currency context where available
- **AND** adjustment transactions are excluded from income and expense totals

### Requirement: Currency Service Behavior
The system SHALL provide service-layer operations for rate lookup, conversion, listing, addition, update, removal, and base-currency access.

#### Scenario: Convert using configured rates
- **WHEN** a service caller converts an amount from a configured foreign currency to the configured base currency
- **THEN** the service returns the converted integer amount using the configured rate

#### Scenario: Base currency conversion is identity
- **WHEN** a service caller converts an amount from the configured base currency to the same base currency
- **THEN** the service returns the original amount without requiring a configured rate entry

#### Scenario: Missing rate returns actionable error
- **WHEN** a service caller requests a rate for an unconfigured foreign currency
- **THEN** the service returns an error that identifies the missing currency and the command to add it
*** Add File: openspec/changes/multi-currency/specs/core-crud/spec.md
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
*** Add File: openspec/changes/multi-currency/specs/wallet-data-model/spec.md
## MODIFIED Requirements

### Requirement: Transactions
The schema SHALL store every money movement as a transaction with support for expenses, income, transfers, balance adjustments, optional planned-payment linkage, original transaction currency, and optional locked base-currency conversion values.

#### Scenario: Standard transaction stores classification and dates
- **WHEN** an expense or income transaction is inserted
- **THEN** it can reference an account and category
- **AND** it stores type, amount, currency, description, notes, business date, creation timestamp, and update timestamp

#### Scenario: Transfer uses destination account reference
- **WHEN** a transfer transaction is inserted
- **THEN** it stores the source account in `account_id`
- **AND** it stores the destination account in `transfer_to_id`

#### Scenario: Multi-currency transaction stores original and base-currency amounts
- **WHEN** a transaction is recorded for an account whose currency differs from the configured base currency
- **THEN** it stores original `amount` and `currency`
- **AND** it stores converted `base_amount` and `base_currency` values that represent the rate locked at transaction creation time
- **AND** later rate configuration changes do not alter the stored conversion values

#### Scenario: Base-currency transaction omits conversion values
- **WHEN** a transaction is recorded for an account whose currency equals the configured base currency
- **THEN** it stores original `amount` and `currency`
- **AND** it does not require `base_amount` or `base_currency`

#### Scenario: Planned payment transaction stores source linkage
- **WHEN** a transaction is created by paying a planned payment
- **THEN** it stores a marker indicating the transaction is planned
- **AND** it stores the source planned payment identifier

### Requirement: Exchange Rates
The schema SHALL store exchange-rate rows for cached or sourced conversion data while manual rate configuration remains the authoritative source for this change.

#### Scenario: Exchange rate cache stores conversion details
- **WHEN** an exchange rate is inserted
- **THEN** it stores source currency, target currency, rate, source label, and fetched timestamp
- **AND** duplicate rows for the same source currency, target currency, and fetched timestamp are rejected

#### Scenario: Manual configured rate is used for transaction conversion
- **WHEN** an income or expense transaction requires conversion during this change
- **THEN** the system uses the local rate configuration as the authoritative rate source
- **AND** the exchange-rate cache is not required to contain a matching row
