## MODIFIED Requirements

### Requirement: Account List Command
The system SHALL provide `wallet account list` to display accounts with balances in a table or JSON format. The total row SHALL convert all account balances to the configured base currency before summing. Negative balances SHALL be displayed and included in the total without restriction.

#### Scenario: List active accounts with single currency
- **WHEN** the user runs `wallet account list` and all accounts share the configured base currency
- **THEN** the system displays a table of non-archived accounts with ID, name, type, currency, balance, and status
- **AND** shows a total balance row with a base-currency label (e.g., `Total (IDR):`)
- **AND** the total is the sum of raw balances (no conversion needed)
- **AND** orders accounts by sort_order and name

#### Scenario: List active accounts with mixed currencies
- **WHEN** the user runs `wallet account list` with accounts in different currencies (e.g., IDR and USD)
- **THEN** the system converts each non-base-currency balance to the base currency using configured exchange rates
- **AND** sums the converted balances to produce the total
- **AND** displays the total with a base-currency label

#### Scenario: List accounts with missing exchange rate
- **WHEN** the user runs `wallet account list` and one account's currency has no configured exchange rate
- **THEN** the system prints a warning listing the missing currency codes
- **AND** excludes the unconvertible account from the total
- **AND** still displays the account row normally with its raw balance

#### Scenario: List accounts with negative balances
- **WHEN** the user runs `wallet account list` and one or more accounts have negative balances
- **THEN** the system displays negative balances in the table with the negative sign prefix
- **AND** includes negative balances in the converted total (which may itself become negative)

#### Scenario: List all accounts including archived
- **WHEN** the user runs `wallet account list --all`
- **THEN** the system includes archived accounts in the table
- **AND** archived accounts display status as "archived"
- **AND** archived account balances are included in the currency-converted total

#### Scenario: List with JSON output
- **WHEN** the user runs `wallet account list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the list of accounts with balances
- **AND** no converted total is included in the JSON output

#### Scenario: No accounts message
- **WHEN** the user runs `wallet account list` and no active accounts exist
- **THEN** the system prints "No accounts found."

## ADDED Requirements

### Requirement: Negative Balance Allowance
The system SHALL allow account balances to be negative at all times without restriction, to support credit cards, loans, overdrafts, and debt tracking.

#### Scenario: Expense exceeds balance
- **WHEN** a user records an expense that exceeds the account's current balance
- **THEN** the system creates the transaction and decrements the account balance
- **AND** the resulting negative balance is stored without error

#### Scenario: Adjust to negative balance
- **WHEN** the user runs `wallet adjust <account> <negative-amount>`
- **THEN** the system sets the account balance to the specified negative value

#### Scenario: Transfer with insufficient funds
- **WHEN** a user transfers an amount exceeding the source account balance
- **THEN** the system warns but allows the transfer
- **AND** the source account balance may become negative
