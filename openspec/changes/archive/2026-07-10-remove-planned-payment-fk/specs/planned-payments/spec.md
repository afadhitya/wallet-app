## MODIFIED Requirements

### Requirement: Planned Payment Fulfillment
The system SHALL provide `wallet bill pay <id>` to manually fulfill a planned payment by creating a standard expense transaction and updating the planned payment state.

#### Scenario: Pay recurring planned payment
- **WHEN** the user runs `wallet bill pay 1` for an active recurring planned payment
- **THEN** the system creates an expense transaction for the planned payment amount, account, category, currency, and payment date
- **AND** decreases the account balance by the transaction amount
- **AND** advances the planned payment next due date to the next occurrence
- **AND** prints the created transaction identifier and next due date

#### Scenario: Pay one-time planned payment
- **WHEN** the user runs `wallet bill pay 3` for an active one-time planned payment
- **THEN** the system creates an expense transaction
- **AND** archives the one-time planned payment so it no longer appears in default active or due views

#### Scenario: Pay with override values
- **WHEN** the user runs `wallet bill pay 1 --date 2026-07-14 --amount 100000`
- **THEN** the system creates the transaction with the overridden transaction date and amount
- **AND** still advances or archives the planned payment according to its schedule type

#### Scenario: Reject paused planned payment fulfillment
- **WHEN** the user runs `wallet bill pay 1` for a paused planned payment
- **THEN** the system exits with a non-zero status
- **AND** does not create a transaction
- **AND** reports that the bill is paused
