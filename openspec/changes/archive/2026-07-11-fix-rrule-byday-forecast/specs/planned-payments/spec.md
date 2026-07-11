## MODIFIED Requirements

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
