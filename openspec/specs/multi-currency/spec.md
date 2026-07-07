# Multi-Currency

## Purpose

TBD

## Requirements

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
The system SHALL present mixed-currency transaction and report output with base-currency totals while preserving original-currency context. Account listing SHALL convert foreign-currency balances to the base currency for the aggregate total.

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

#### Scenario: Account list total in base currency
- **WHEN** the user runs `wallet account list` with accounts in multiple currencies
- **THEN** the total at the bottom of the table is denominated in the configured base currency
- **AND** each non-base-currency account balance is converted using the configured exchange rate before summing
- **AND** a warning is emitted if any account's currency lacks a configured rate

#### Scenario: Account list total when all accounts share base currency
- **WHEN** the user runs `wallet account list` and all accounts use the configured base currency
- **THEN** the total is the direct sum of raw balances with a base-currency label

#### Scenario: Forecast aggregated start balance in base currency
- **WHEN** the user runs `wallet forecast` (without `--account`) and holds accounts in multiple currencies
- **THEN** the starting balance equals the sum of each account's balance converted to the configured base currency using the configured exchange rates
- **AND** accounts whose currency matches the base currency contribute their raw balance directly

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
