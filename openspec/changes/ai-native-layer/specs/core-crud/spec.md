## ADDED Requirements

### Requirement: Core CRUD AI-Native JSON Output
The system SHALL render core CRUD command results and failures through the shared AI-native JSON envelope when `--json` is supplied.

#### Scenario: Add transaction returns envelope JSON
- **WHEN** the user runs `wallet add expense 35000 "Lunch at Warung" -c food -a bca --json`
- **THEN** the system records the transaction normally
- **AND** writes a JSON envelope with `success: true`
- **AND** `data` contains the created transaction fields including ID, type, amount, description, category, account, tags, and planned-payment state where applicable
- **AND** `meta.command` identifies the add command

#### Scenario: List transactions returns envelope JSON
- **WHEN** the user runs `wallet list --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.transactions` contains the listed transactions
- **AND** `data.total` contains the listed transaction total
- **AND** `data.count` contains the number of listed transactions

#### Scenario: Transfer returns envelope JSON
- **WHEN** the user runs `wallet add transfer 100000 --from bca --to gopay --json`
- **THEN** the system records the transfer normally
- **AND** writes a JSON envelope with `success: true`
- **AND** `data` identifies the transfer transaction, source account, destination account, amount, and any warnings

#### Scenario: Core CRUD errors return envelope JSON
- **WHEN** the user runs a core CRUD command with `--json` and references a missing category, account, tag, or transaction
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the missing or invalid resource type
- **AND** `error.message` describes the failure without table formatting
