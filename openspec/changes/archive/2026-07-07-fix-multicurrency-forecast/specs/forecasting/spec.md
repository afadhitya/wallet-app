## MODIFIED Requirements

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
