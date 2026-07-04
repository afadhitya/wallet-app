## ADDED Requirements

### Requirement: Planned Payment Service And Query Support
The system SHALL implement service-layer operations and sqlc-backed queries for planned payment creation, listing, due filtering, fulfillment, skipping, pausing, resuming, editing, deleting, and recurrence calculation.

#### Scenario: Services validate and persist planned payment operations
- **WHEN** CLI commands invoke planned-payment service methods
- **THEN** the services validate domain inputs before writing
- **AND** use sqlc-generated queries for database access
- **AND** return typed results or clear errors for the CLI to render

#### Scenario: Pay operation updates transaction and planned payment state
- **WHEN** the planned-payment service pays a bill
- **THEN** it creates the linked transaction
- **AND** updates affected account balance through existing transaction balance behavior
- **AND** advances or archives the planned payment in the same logical operation

#### Scenario: Due filters ignore inactive and paused payments
- **WHEN** the planned-payment service lists due payments
- **THEN** it excludes paused planned payments
- **AND** excludes archived planned payments

### Requirement: Planned Payment CLI Commands
The system SHALL expose planned-payment workflows through a `wallet bill` command group.

#### Scenario: Bill command group is available
- **WHEN** the user runs `wallet bill --help`
- **THEN** the system shows subcommands for add, list, due, pay, skip, pause, resume, edit, and rm

#### Scenario: Bill commands render text and JSON output
- **WHEN** the user runs a bill command that supports `--json`
- **THEN** the system renders JSON output instead of table or prose output

#### Scenario: Bill commands report missing records
- **WHEN** the user runs a bill command with an identifier that does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that the bill was not found

### Requirement: Planned Payment Testing
The system SHALL include unit and integration tests for planned payment service and CLI behavior with deterministic local databases and SHALL satisfy the repository's 100% Go coverage gate after approved generated-code and documented OS/infrastructure exclusions are applied.

#### Scenario: Service tests cover recurrence and state transitions
- **WHEN** planned-payment service tests run
- **THEN** they verify creation validation, due filtering, pay, skip, pause, resume, edit, delete, and recurrence edge cases against an isolated SQLite database

#### Scenario: CLI integration tests cover bill workflows
- **WHEN** CLI integration tests execute planned-payment commands
- **THEN** they verify exit codes, stable output content, JSON output where supported, and database side effects

#### Scenario: Coverage gate passes for planned payment implementation
- **WHEN** GitHub Actions runs the repository coverage check after planned payments are implemented
- **THEN** total included Go test coverage remains exactly `100%`
- **AND** generated sqlc code and documented OS/infrastructure failure branches remain excluded by the approved coverage policy
