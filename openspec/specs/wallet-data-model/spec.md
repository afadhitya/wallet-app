# Wallet Data Model

## Purpose

TBD - Define the purpose of the wallet data model specification.

## Requirements

### Requirement: Database Initialization
The application SHALL provide a way to initialize a SQLite wallet database with the approved core schema through embedded migrations.

#### Scenario: Initialize empty wallet database
- **WHEN** the database initialization runs against an empty SQLite database
- **THEN** it creates the core wallet tables for accounts, categories, tags, transactions, transaction tags, budgets, budget categories, budget tags, planned payments, and exchange rates
- **AND** foreign key enforcement is enabled for the connection used by the application
- **AND** the initial schema is applied from an embedded SQL migration file

#### Scenario: Track applied schema version
- **WHEN** the database initialization successfully applies a migration
- **THEN** it records the applied migration version so subsequent startup runs do not reapply the same migration

### Requirement: Account Storage
The schema SHALL store wallet accounts as money storage entities with integer balances in minor units.

#### Scenario: Account record captures wallet storage details
- **WHEN** an account is inserted
- **THEN** the schema supports name, type, currency, balance, archived state, sort order, creation timestamp, and update timestamp
- **AND** the balance is stored as an integer minor-unit value

### Requirement: Category Hierarchy
The schema SHALL support a two-level category hierarchy with system seed categories.

#### Scenario: Default categories are available after initialization
- **WHEN** a new wallet database is initialized
- **THEN** default system parent and child categories exist for common expense and income classifications

#### Scenario: Child category references a parent
- **WHEN** a child category is inserted with a parent category
- **THEN** the child category references the parent through `parent_id`
- **AND** categories do not require recursive hierarchy support beyond parent and child rows

### Requirement: Tags
The schema SHALL support unique freeform tags for cross-cutting transaction and budget labels.

#### Scenario: Duplicate tag names are rejected
- **WHEN** two tags are inserted with the same name
- **THEN** the second insert fails due to the unique tag name constraint

### Requirement: Transactions
The schema SHALL store every money movement as a transaction with support for expenses, income, transfers, balance adjustments, and optional planned-payment linkage.

#### Scenario: Standard transaction stores classification and dates
- **WHEN** an expense or income transaction is inserted
- **THEN** it can reference an account and category
- **AND** it stores type, amount, currency, description, notes, business date, creation timestamp, and update timestamp

#### Scenario: Transfer uses destination account reference
- **WHEN** a transfer transaction is inserted
- **THEN** it stores the source account in `account_id`
- **AND** it stores the destination account in `transfer_to_id`

#### Scenario: Multi-currency transaction stores original and account-currency amounts
- **WHEN** a transaction is recorded in a currency different from the account currency
- **THEN** it stores original `amount` and `currency`
- **AND** it can store converted `base_amount` and `base_currency`

#### Scenario: Planned payment transaction stores source linkage
- **WHEN** a transaction is created by paying a planned payment
- **THEN** it stores a marker indicating the transaction is planned
- **AND** it stores the source planned payment identifier

### Requirement: Transaction Tags
The schema SHALL support many-to-many tags on transactions.

#### Scenario: Transaction can have multiple tags
- **WHEN** tag rows are associated with a transaction through `transaction_tags`
- **THEN** multiple tags can be linked to the same transaction
- **AND** deleting the transaction removes its tag links

### Requirement: Budget Snapshots
The schema SHALL store budgets as per-period snapshots with category and tag targets.

#### Scenario: Budget stores period and limit
- **WHEN** a budget is inserted
- **THEN** it stores name, amount, currency, type, period start, period end, notification percentage, active state, creation timestamp, and update timestamp

#### Scenario: Budget targets categories and tags
- **WHEN** a budget is linked through `budget_categories` or `budget_tags`
- **THEN** the budget can target multiple categories, multiple tags, or both
- **AND** deleting the budget removes its target links

### Requirement: Planned Payments
The schema SHALL store planned income and expense payments with precomputed due dates, recurrence state, active state, and pause state.

#### Scenario: Planned payment stores recurrence details
- **WHEN** a planned payment is inserted
- **THEN** it stores account, optional category, type, amount, currency, name, recurrence, optional recurrence rule, start date, next due date, active state, paused state, creation timestamp, and update timestamp

#### Scenario: Archived planned payment remains linkable
- **WHEN** a one-time planned payment is paid and archived
- **THEN** existing transactions can still reference the archived planned payment
- **AND** default active planned-payment queries exclude it

### Requirement: Exchange Rates
The schema SHALL store cached exchange rates for currency conversion.

#### Scenario: Exchange rate cache stores conversion details
- **WHEN** an exchange rate is inserted
- **THEN** it stores source currency, target currency, rate, source label, and fetched timestamp
- **AND** duplicate rows for the same source currency, target currency, and fetched timestamp are rejected
