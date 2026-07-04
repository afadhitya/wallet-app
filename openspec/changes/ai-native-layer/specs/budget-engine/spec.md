## ADDED Requirements

### Requirement: Budget AI-Native JSON Output
The system SHALL render budget command results and failures through the shared AI-native JSON envelope when `--json` is supplied.

#### Scenario: Budget list returns envelope JSON
- **WHEN** the user runs `wallet budget list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.budgets` contains budget rows with ID, name, limit, spent amount, remaining amount, period fields, target fields, notification threshold, and active state
- **AND** the response does not include table formatting

#### Scenario: Budget check returns envelope JSON
- **WHEN** the user runs `wallet budget check --all --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.budgets` contains each checked budget's ID, name, limit, spent, remaining, percent used, status, period start, and period end

#### Scenario: Budget errors return envelope JSON
- **WHEN** the user runs a budget command with `--json` and provides invalid input or references a missing budget
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies invalid budget input or `BUDGET_NOT_FOUND`
- **AND** `error.message` describes the failure without table formatting
