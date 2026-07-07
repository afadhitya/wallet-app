## MODIFIED Requirements

### Requirement: COMMANDS.md groups commands by domain

The `skill/COMMANDS.md` file SHALL organize commands into domain-grouped sections (Transaction, Account, Category, Tag, Budget, Bill, Forecast, Report, Rate, Init, System) rather than a flat alphabetical list. The System section SHALL document `wallet version [--check]` and `wallet update [--force]`.

#### Scenario: Agent needs account-related commands

- **WHEN** an AI agent searches for account commands
- **THEN** the agent SHALL find all account subcommands (`account add`, `account list`, `account edit`, `account archive`) grouped under a single "Account" heading

#### Scenario: Agent needs version or update commands

- **WHEN** an AI agent searches for version or update commands
- **THEN** the agent SHALL find `wallet version` (with `--check` flag) and `wallet update` (with `--force` flag) grouped under a "System" heading

### Requirement: ERRORS.md includes recovery patterns

The `skill/ERRORS.md` file SHALL document each error code with its meaning, cause, and suggested recovery action, plus common recovery patterns as a reference table. Update-related error codes (`UPDATE_NETWORK_ERROR`, `UPDATE_PERMISSION_ERROR`, `UPDATE_ALREADY_LATEST`, `UPDATE_FAILED`, `UPDATE_CHECKSUM_MISMATCH`) SHALL be included.

#### Scenario: Agent receives UPDATE_NETWORK_ERROR

- **WHEN** an AI agent receives an `UPDATE_NETWORK_ERROR` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "Check internet connection and retry"

#### Scenario: Agent receives UPDATE_PERMISSION_ERROR

- **WHEN** an AI agent receives an `UPDATE_PERMISSION_ERROR` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "Run with appropriate permissions or reinstall"

#### Scenario: Agent receives UPDATE_ALREADY_LATEST

- **WHEN** an AI agent receives an `UPDATE_ALREADY_LATEST` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "No action needed; force with --force if reinstall is desired"

#### Scenario: Agent receives UPDATE_FAILED

- **WHEN** an AI agent receives an `UPDATE_FAILED` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "Check the error message for details; may require manual reinstall"

#### Scenario: Agent receives UPDATE_CHECKSUM_MISMATCH

- **WHEN** an AI agent receives an `UPDATE_CHECKSUM_MISMATCH` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "Retry update; checksum verification failed"
