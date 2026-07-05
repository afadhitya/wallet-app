## ADDED Requirements

### Requirement: Account Management JSON Output Support
The system SHALL support the global `--json` flag for all `wallet account` subcommands, rendering results and errors through the shared AI-native JSON envelope.

#### Scenario: Account add follows JSON envelope contract
- **WHEN** the user runs `wallet account add "BCA" --json`
- **THEN** the system writes a success envelope with `meta.command` identifying the account add command
- **AND** uses the shared `printSuccessJSON` helper for consistency

#### Scenario: Account list follows JSON envelope contract
- **WHEN** the user runs `wallet account list --json`
- **THEN** the system writes a success envelope with `meta.command` identifying the account list command
- **AND** uses the shared `printSuccessJSON` helper for consistency

#### Scenario: Account edit follows JSON envelope contract
- **WHEN** the user runs `wallet account edit 1 --name "Updated" --json`
- **THEN** the system writes a success envelope with `meta.command` identifying the account edit command

#### Scenario: Account archive follows JSON envelope contract
- **WHEN** the user runs `wallet account archive 1 --force --json`
- **THEN** the system writes a success envelope with `meta.command` identifying the account archive command

#### Scenario: Account errors use structured JSON error codes
- **WHEN** the user runs an account command with `--json` and encounters a `NotFoundError` for an account
- **THEN** the system uses the shared `classifyError` function
- **AND** returns the existing `ACCOUNT_NOT_FOUND` error code
- **AND** includes a suggestion when a likely match is identifiable
