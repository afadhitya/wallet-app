# Planned Payments

## Purpose

TBD - Define the purpose of the planned payments specification.

## Requirements

### Requirement: Planned Payment Creation
The system SHALL provide `wallet bill add` to create active planned payments for recurring bills and one-time future expenses.

#### Scenario: Create monthly planned payment
- **WHEN** the user runs `wallet bill add "Netflix" 149000 --monthly --day 15 -c subscriptions -a bca`
- **THEN** the system creates an active planned payment named `Netflix` for amount `149000`
- **AND** stores the selected account and category references
- **AND** stores monthly recurrence with next due date on the 15th day according to the start date
- **AND** prints a success message or JSON representation according to the output mode

#### Scenario: Create one-time future payment
- **WHEN** the user runs `wallet bill add "Flight to Tokyo" 3000000 -c travel -a bca --from 2026-08-15`
- **THEN** the system creates an active one-time planned payment due on `2026-08-15`
- **AND** stores no recurring schedule beyond the one-time due date

#### Scenario: Reject invalid planned payment input
- **WHEN** the user creates a planned payment with a non-positive amount, missing schedule, missing account, missing category, or invalid recurrence rule
- **THEN** the system exits with a non-zero status
- **AND** does not create a planned payment
- **AND** reports a clear validation error that identifies the invalid field

### Requirement: Planned Payment Listing
The system SHALL provide `wallet bill list` to list planned payments with filters for active, paused, archived, all, and JSON output.

#### Scenario: List active planned payments by default
- **WHEN** the user runs `wallet bill list` without filters
- **THEN** the system lists active and non-paused planned payments ordered by next due date
- **AND** excludes paused and archived planned payments

#### Scenario: List paused planned payments
- **WHEN** the user runs `wallet bill list --paused`
- **THEN** the system lists paused planned payments
- **AND** excludes active non-paused planned payments unless `--active` is also supplied

#### Scenario: Include archived planned payments
- **WHEN** the user runs `wallet bill list --all`
- **THEN** the system includes active, paused, and archived planned payments

### Requirement: Due Planned Payment View
The system SHALL provide `wallet bill due` to show unpaid active planned payments due within a selected window.

#### Scenario: Show current-month due payments by default
- **WHEN** the user runs `wallet bill due` without filters
- **THEN** the system lists active non-paused planned payments with next due dates within the current month
- **AND** includes a total due amount in text output

#### Scenario: Show overdue payments
- **WHEN** the user runs `wallet bill due --overdue`
- **THEN** the system lists active non-paused planned payments whose next due date is before the current date
- **AND** indicates how long each listed planned payment is overdue in text output

#### Scenario: Show next N days
- **WHEN** the user runs `wallet bill due --next 14`
- **THEN** the system lists active non-paused planned payments due within the next 14 days

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

### Requirement: Planned Payment Skip
The system SHALL provide `wallet bill skip <id>` to skip exactly one occurrence of a recurring planned payment without creating a transaction.

#### Scenario: Skip recurring planned payment
- **WHEN** the user runs `wallet bill skip 2` for an active recurring planned payment
- **THEN** the system advances the planned payment next due date to the next occurrence
- **AND** creates no transaction
- **AND** prints the updated next due date

#### Scenario: Reject one-time skip
- **WHEN** the user runs `wallet bill skip 3` for a one-time planned payment
- **THEN** the system exits with a non-zero status
- **AND** does not change the planned payment
- **AND** reports that one-time planned payments cannot be skipped

### Requirement: Planned Payment State Management
The system SHALL provide commands to pause, resume, edit, and remove planned payments.

#### Scenario: Pause planned payment
- **WHEN** the user runs `wallet bill pause 1`
- **THEN** the system marks planned payment `1` as paused
- **AND** excludes it from default list and due results

#### Scenario: Resume planned payment
- **WHEN** the user runs `wallet bill resume 1`
- **THEN** the system marks planned payment `1` as not paused
- **AND** keeps or recalculates its next due date according to the service rules

#### Scenario: Edit planned payment fields
- **WHEN** the user runs `wallet bill edit 1 --amount 169000 --name "Netflix Premium" --category subscriptions --account bca --day 20`
- **THEN** the system updates only the supplied planned payment fields
- **AND** recalculates the next due date when the due day changes

#### Scenario: Remove planned payment
- **WHEN** the user runs `wallet bill rm 1`
- **THEN** the system deletes or deactivates planned payment `1` according to referential constraints
- **AND** excludes it from future default list and due results

### Requirement: Recurrence Calculation
The system SHALL calculate next due dates deterministically for daily, weekly, monthly, yearly, custom, and one-time schedules.

#### Scenario: Advance simple recurrence
- **WHEN** a daily, weekly, monthly, or yearly planned payment is paid or skipped
- **THEN** the system advances its next due date by one day, one week, one month, or one year respectively

#### Scenario: Clamp monthly recurrence to end of month
- **WHEN** a monthly planned payment due on January 31 is advanced into February
- **THEN** the system sets the next due date to the last valid day of February

#### Scenario: Advance custom WEEKLY recurrence with BYDAY
- **WHEN** a custom WEEKLY planned payment with `BYDAY=TU,WE` (e.g., `FREQ=WEEKLY;BYDAY=TU,WE`) is paid, skipped, or expanded in a forecast
- **THEN** the system calculates the next due date as the next calendar date matching one of the specified weekdays
- **AND** advances from the current due date + 1 day to find the earliest matching day of the week
- **AND** a plain `FREQ=WEEKLY` (no `BYDAY`) continues to advance by exactly 7 days

#### Scenario: Reject invalid custom recurrence rule
- **WHEN** the user creates or edits a planned payment with an invalid RRULE
- **THEN** the system exits with a non-zero status
- **AND** reports the recurrence rule validation error

### Requirement: Planned Payment JSON Output
The system SHALL support AI-native JSON output for planned-payment commands when `--json` is supplied and SHALL use the shared JSON response envelope for successes and failures.

#### Scenario: Render planned payment JSON
- **WHEN** the user runs a planned-payment command with `--json`
- **THEN** the system writes a machine-readable JSON response containing `success: true`
- **AND** the response contains `data` with command result fields
- **AND** the response contains `meta.command` and `meta.timestamp`
- **AND** the response does not include table formatting in the response

#### Scenario: Bill due returns envelope JSON
- **WHEN** the user runs `wallet bill due --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.due` contains active, unpaused planned payments due in the selected window
- **AND** `data.total_due` contains the total due amount
- **AND** `data.count` contains the number of due payments

#### Scenario: Planned payment errors return envelope JSON
- **WHEN** the user runs a planned-payment command with `--json` and references a missing, paused, already-paid, or invalid bill
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the planned-payment failure condition
- **AND** `error.message` describes the failure without table formatting
