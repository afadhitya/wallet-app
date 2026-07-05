# Account List Converted Balance

## Purpose

TBD

## Requirements

### Requirement: Account List Shows Per-Row Converted Balance
The system SHALL display a "Converted" column in the `wallet account list` table that shows each account's balance converted to the configured base currency.

#### Scenario: Non-base-currency account shows converted balance
- **WHEN** the user runs `wallet account list` and an account's currency differs from the configured base currency
- **AND** a configured exchange rate exists for that currency
- **THEN** the account row includes a "Converted" column showing the balance multiplied by the exchange rate
- **AND** the converted value is formatted with the base currency prefix and thousand separators

#### Scenario: Base-currency account shows no converted balance
- **WHEN** the user runs `wallet account list` and an account's currency equals the configured base currency
- **THEN** the account row shows `-` in the "Converted" column

#### Scenario: Account with missing rate shows no converted balance
- **WHEN** the user runs `wallet account list` and an account's currency differs from the base currency
- **AND** no configured exchange rate exists for that currency
- **THEN** the account row shows `-` in the "Converted" column
- **AND** a warning about the missing rate is still emitted on stderr

#### Scenario: Table header includes Converted column
- **WHEN** the user runs `wallet account list`
- **THEN** the table header includes a "Converted" column between "Balance" and "Status"
- **AND** the header separator row includes the corresponding dashes

#### Scenario: Converted column for negative balances
- **WHEN** the user runs `wallet account list` and a non-base-currency account has a negative balance
- **THEN** the converted balance is also negative
- **AND** the converted value is formatted with the negative sign prefix

#### Scenario: Converted column unaffected by JSON output
- **WHEN** the user runs `wallet account list --json`
- **THEN** the JSON output contains the raw account data without a converted balance field
- **AND** the table-specific "Converted" column does not appear in JSON output
