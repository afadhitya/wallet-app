# AI-Native CLI

## Purpose

TBD

## Requirements

### Requirement: Global JSON Output Envelope
The system SHALL support a global `--json` flag for every wallet command and SHALL write successful machine-readable responses with a consistent JSON envelope.

#### Scenario: Render successful JSON envelope
- **WHEN** the user runs any wallet command with `--json`
- **THEN** the system writes valid JSON to standard output
- **AND** the top-level response contains `success: true`
- **AND** the top-level response contains `data` with command-specific result fields
- **AND** the top-level response contains `meta.command` identifying the executed command
- **AND** the top-level response contains `meta.timestamp` as an RFC3339 UTC timestamp
- **AND** the response does not include table formatting or prose-only output outside JSON

#### Scenario: Preserve default text output
- **WHEN** the user runs a wallet command without `--json`
- **THEN** the system renders the command's existing human-readable text, table, or prose output
- **AND** the system does not wrap the response in a JSON envelope

### Requirement: Structured JSON Errors
The system SHALL render command failures as structured JSON errors when `--json` is supplied.

#### Scenario: Render validation error JSON
- **WHEN** the user runs a wallet command with `--json` and invalid input
- **THEN** the system exits with a non-zero status
- **AND** the response contains `success: false`
- **AND** the response contains `error.code` with a stable machine-readable code
- **AND** the response contains `error.message` with a human-readable explanation
- **AND** the response MAY contain `error.suggestion` when a useful correction is available

#### Scenario: Render not-found suggestion JSON
- **WHEN** the user runs a wallet command with `--json` and references a missing category, account, tag, budget, or bill
- **THEN** the system exits with a non-zero status
- **AND** the response contains a specific not-found error code for the missing resource type
- **AND** the response includes a suggestion when the system can identify a likely match

### Requirement: Wallet Agent Skill
The system SHALL include a repository-local wallet agent skill that instructs AI agents how to invoke wallet CLI commands with JSON output.

#### Scenario: Agent skill is available in the repository
- **WHEN** the repository contents are inspected
- **THEN** `skill/SKILL.md` exists in the repository root
- **AND** the skill metadata identifies the wallet finance use case
- **AND** the skill describes trigger words for expenses, income, budgets, bills, forecasts, and wallet usage

#### Scenario: Agent maps natural language to JSON CLI command
- **WHEN** an AI agent follows the wallet skill for a user finance request
- **THEN** the agent maps the request to a wallet CLI command
- **AND** the agent appends `--json` to the command
- **AND** the agent parses the JSON response envelope before formatting a friendly reply

#### Scenario: Agent does not auto-create missing tags
- **WHEN** the user references a tag that does not exist
- **THEN** the skill instructs the agent not to auto-create the tag implicitly
- **AND** the agent asks the user or directs them to create the tag first
