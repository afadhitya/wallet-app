## ADDED Requirements

### Requirement: Account Add Command
The system SHALL provide `wallet account add <name>` to create a new account with optional type and currency.

#### Scenario: Add account with defaults
- **WHEN** the user runs `wallet account add "BCA"`
- **THEN** the system creates an account named "BCA" with type "checking" and currency "IDR"
- **AND** prints a success message or JSON representation according to the output mode

#### Scenario: Add account with explicit type and currency
- **WHEN** the user runs `wallet account add "GoPay" --type ewallet --currency IDR`
- **THEN** the system creates an account with the specified type and currency
- **AND** returns the created account details including ID

#### Scenario: Reject duplicate account name
- **WHEN** the user runs `wallet account add "BCA"` and an account named "BCA" already exists (case-insensitive)
- **THEN** the system exits with a non-zero status
- **AND** reports the duplicate name error

#### Scenario: Reject invalid account type
- **WHEN** the user runs `wallet account add "Test" --type invalid`
- **THEN** the system exits with a non-zero status
- **AND** reports the invalid type with valid options

#### Scenario: Reject empty account name
- **WHEN** the user runs `wallet account add ""`
- **THEN** the system exits with a non-zero status
- **AND** reports that the account name is required

### Requirement: Account List Command
The system SHALL provide `wallet account list` to display accounts with balances in a table or JSON format.

#### Scenario: List active accounts
- **WHEN** the user runs `wallet account list`
- **THEN** the system displays a table of non-archived accounts with ID, name, type, currency, balance, and status
- **AND** shows a total balance row at the bottom
- **AND** orders accounts by sort_order and name

#### Scenario: List all accounts including archived
- **WHEN** the user runs `wallet account list --all`
- **THEN** the system includes archived accounts in the table
- **AND** archived accounts display status as "archived"

#### Scenario: List with JSON output
- **WHEN** the user runs `wallet account list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the list of accounts with balances

#### Scenario: No accounts message
- **WHEN** the user runs `wallet account list` and no active accounts exist
- **THEN** the system prints "No accounts found."

### Requirement: Account Edit Command
The system SHALL provide `wallet account edit <id>` to update an existing account's name, type, and sort order.

#### Scenario: Edit account name
- **WHEN** the user runs `wallet account edit 1 --name "BCA Main"`
- **THEN** the system updates account 1's name to "BCA Main"
- **AND** preserves all other account fields (type, currency, balance)
- **AND** prints a success message or JSON representation

#### Scenario: Edit account type
- **WHEN** the user runs `wallet account edit 1 --type savings`
- **THEN** the system updates account 1's type to "savings"
- **AND** does not change the account name or currency

#### Scenario: Reject edit of non-existent account
- **WHEN** the user runs `wallet account edit 99 --name "Ghost"` and account 99 does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that the account was not found

#### Scenario: Reject empty name on edit
- **WHEN** the user runs `wallet account edit 1 --name ""`
- **THEN** the system exits with a non-zero status
- **AND** reports that the account name cannot be empty

#### Scenario: Reject edit with no changes
- **WHEN** the user runs `wallet account edit 1` with no flags
- **THEN** the system exits with a non-zero status
- **AND** reports that at least one field must be specified

### Requirement: Account Archive Command
The system SHALL provide `wallet account archive <id>` to soft-delete an account with an optional confirmation bypass.

#### Scenario: Archive account with confirmation
- **WHEN** the user runs `wallet account archive 1`
- **THEN** the system prompts for confirmation
- **AND** archives the account on confirmation
- **AND** does not archive the account if the user declines

#### Scenario: Archive account with force flag
- **WHEN** the user runs `wallet account archive 1 --force`
- **THEN** the system archives the account without prompting

#### Scenario: Warn on non-zero balance
- **WHEN** the user runs `wallet account archive 1` and the account has a non-zero balance
- **THEN** the system displays a warning with the balance amount
- **AND** still proceeds with archive after confirmation or `--force`

#### Scenario: Reject archive of non-existent account
- **WHEN** the user runs `wallet account archive 99` and account 99 does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that the account was not found

#### Scenario: Archived account excluded from default list
- **WHEN** an account is archived
- **THEN** the account no longer appears in `wallet account list` (without `--all`)
- **AND** is still visible with `wallet account list --all`

### Requirement: Account Management JSON Output
The system SHALL render account command results and failures through the shared AI-native JSON envelope when `--json` is supplied.

#### Scenario: Add account returns envelope JSON
- **WHEN** the user runs `wallet account add "BCA" --type checking --currency IDR --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the created account fields including ID, name, type, currency, balance, and timestamps

#### Scenario: List accounts returns envelope JSON
- **WHEN** the user runs `wallet account list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the array of account objects with balances

#### Scenario: Edit account returns envelope JSON
- **WHEN** the user runs `wallet account edit 1 --name "Updated" --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the updated account fields

#### Scenario: Archive account returns envelope JSON
- **WHEN** the user runs `wallet account archive 1 --force --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains a confirmation that the account was archived

#### Scenario: Account errors return envelope JSON
- **WHEN** the user runs an account command with `--json` and encounters an error (not found, duplicate name, invalid type)
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the error type (e.g., `ACCOUNT_NOT_FOUND`, `VALIDATION_ERROR`)
- **AND** `error.message` describes the failure
