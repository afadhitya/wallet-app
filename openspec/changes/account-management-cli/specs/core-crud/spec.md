## ADDED Requirements

### Requirement: Account Management Commands
The system SHALL provide account lifecycle management through a `wallet account` command group with `add`, `list`, `edit`, and `archive` subcommands.

#### Scenario: Account command group is available
- **WHEN** the user runs `wallet account --help`
- **THEN** the system shows subcommands for add, list, edit, and archive

#### Scenario: Account commands follow existing core CRUD patterns
- **WHEN** the user invokes any account subcommand
- **THEN** the system validates inputs before persisting changes
- **AND** uses the existing account service methods for all operations
- **AND** renders text tables or JSON output matching the existing core CRUD command style
- **AND** returns typed errors (not found, validation, duplicate) matching the existing error classification

#### Scenario: Account commands require initialized wallet
- **WHEN** the user runs an account command before running `wallet init`
- **THEN** the system exits with a non-zero status
- **AND** reports that the database is not initialized
