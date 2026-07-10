## MODIFIED Requirements

### Requirement: Transactions
The schema SHALL store every money movement as a transaction with support for expenses, income, transfers, balance adjustments, original transaction currency, and optional locked base-currency conversion values.

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

## REMOVED Requirements

### Requirement: Planned payment transaction stores source linkage
**Reason**: The `planned_payment_id` and `is_planned` columns are being removed from the `transactions` table. No report, budget, or CLI query reads these columns.
**Migration**: Transactions created by `bill pay` are now plain records with no planned payment linkage. No data migration needed — existing values are discarded.

### Requirement: Archived planned payment remains linkable
**Reason**: With the FK removed, there is no longer any linkage between transactions and planned payments. The concept of "linkable" is moot.
**Migration**: No action needed. Archived planned payments continue to exist independently of transactions.
