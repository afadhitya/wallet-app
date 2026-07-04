## MODIFIED Requirements

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

### Requirement: Planned Payments
The schema SHALL store planned income and expense payments with precomputed due dates, recurrence state, active state, and pause state.

#### Scenario: Planned payment stores recurrence details
- **WHEN** a planned payment is inserted
- **THEN** it stores account, optional category, type, amount, currency, name, recurrence, optional recurrence rule, start date, next due date, active state, paused state, creation timestamp, and update timestamp

#### Scenario: Archived planned payment remains linkable
- **WHEN** a one-time planned payment is paid and archived
- **THEN** existing transactions can still reference the archived planned payment
- **AND** default active planned-payment queries exclude it
