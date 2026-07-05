## MODIFIED Requirements

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
