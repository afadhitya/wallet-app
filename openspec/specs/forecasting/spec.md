# Forecasting

## Purpose

Provide balance and bill forecasting capabilities so users can project their financial position forward in time based on planned payments.

## Requirements

### Requirement: Balance Forecast Command
The system SHALL provide `wallet forecast` to project future balances from active, unpaused planned payments.

#### Scenario: Show default one-month forecast
- **WHEN** the user runs `wallet forecast` without flags
- **THEN** the system projects balances for the next 1 month
- **AND** converts each non-base-currency account balance to the base currency using configured exchange rates before summing
- **AND** uses the base-currency sum of all non-archived account balances as the starting balance
- **AND** includes active, unpaused planned income and expense payments due within the forecast horizon
- **AND** excludes paused, archived, and undated planned payments
- **AND** prints projected income, projected expenses, net movement, ending balance, and the planned payments used for the projection

#### Scenario: Show multi-month forecast
- **WHEN** the user runs `wallet forecast --months 3`
- **THEN** the system projects balances for 3 monthly buckets
- **AND** each month starts from the previous month ending balance
- **AND** each month includes projected income, projected expenses, net movement, and ending balance
- **AND** recurring planned payments contribute each occurrence due in the requested horizon
- **AND** monthly projected income and expenses aggregate planned payment amounts converted to the base currency

#### Scenario: Skip planned payments with missing exchange rate
- **WHEN** the user runs `wallet forecast` and a planned payment's account currency lacks a configured exchange rate
- **THEN** the system excludes that planned payment from the projection
- **AND** logs a warning identifying the skipped payment and missing currency

#### Scenario: Reject invalid forecast horizon
- **WHEN** the user runs `wallet forecast --months 0`
- **THEN** the system exits with a non-zero status
- **AND** reports that the forecast horizon must be positive

### Requirement: Account-Scoped Forecast
The system SHALL allow balance forecasts to be limited to one account with `--account` or `-a`.

#### Scenario: Forecast one account
- **WHEN** the user runs `wallet forecast --account bca --months 2`
- **THEN** the system resolves `bca` as an account name or identifier
- **AND** projects balances using only that account's current balance and planned payments
- **AND** excludes planned payments for other accounts

#### Scenario: Reject unknown account
- **WHEN** the user runs `wallet forecast --account unknown`
- **THEN** the system exits with a non-zero status
- **AND** reports that the account was not found

### Requirement: Forecast Breakdown
The system SHALL include category-level and bill-level details in balance forecasts when planned payment data is available.

#### Scenario: Group forecast by category
- **WHEN** the forecast includes planned payments with categories
- **THEN** text output includes a category breakdown of projected expenses
- **AND** the breakdown totals planned expense amounts by category within the forecast horizon

#### Scenario: Show planned payment details
- **WHEN** the forecast includes planned payments due in the horizon
- **THEN** output includes each planned payment name, amount, due date, and account context where applicable

### Requirement: Bills Forecast Command
The system SHALL provide `wallet forecast bills` to show upcoming planned expense impact over a forecast horizon.

#### Scenario: Show upcoming bills
- **WHEN** the user runs `wallet forecast bills`
- **THEN** the system lists active, unpaused planned expense payments due within the default 2-month horizon
- **AND** orders bills by due date ascending
- **AND** includes each bill date, name, amount, and running total
- **AND** prints the total upcoming bill amount

#### Scenario: Show bills for custom horizon
- **WHEN** the user runs `wallet forecast bills --months 2`
- **THEN** the system lists planned expense occurrences due within the next 2 months
- **AND** recurring planned expenses contribute each occurrence due in that horizon

### Requirement: Forecast JSON Output
The system SHALL support machine-readable AI-native JSON output for forecast commands when global `--json` is supplied and SHALL use the shared JSON response envelope.

#### Scenario: Render balance forecast JSON
- **WHEN** the user runs `wallet forecast --months 3 --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the forecast horizon, forecast rows, planned payment details, totals, and warnings
- **AND** `meta.command` identifies the forecast command
- **AND** the response does not include table formatting in the response

#### Scenario: Render bills forecast JSON
- **WHEN** the user runs `wallet forecast bills --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains bill rows, running totals, total amount, count, and horizon
- **AND** `meta.command` identifies the forecast bills command
- **AND** the response does not include table formatting in the response

#### Scenario: Forecast errors return envelope JSON
- **WHEN** the user runs a forecast command with `--json` and provides an invalid horizon or unknown account
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the invalid horizon or missing account
- **AND** `error.message` describes the failure without table formatting

### Requirement: Forecast Warnings and Empty State
The system SHALL communicate forecast limitations and warning states without failing successful forecast commands.

#### Scenario: No planned payments found
- **WHEN** the user runs a forecast command and no matching planned payments exist
- **THEN** the system exits with status 0
- **AND** reports that no planned payments were found
- **AND** tells the user forecasts are based on planned payments

#### Scenario: Projected negative balance
- **WHEN** a forecast month has a projected ending balance below zero
- **THEN** the system exits with status 0
- **AND** includes a warning identifying the affected month and account
