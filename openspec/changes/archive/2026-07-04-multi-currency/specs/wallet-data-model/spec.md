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
