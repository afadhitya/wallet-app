## ADDED Requirements

### Requirement: Command reference is concise signature inventory

The `skill/COMMANDS.md` SHALL be a concise command inventory listing only command signatures grouped by domain, without parameter tables, flag descriptions, or JSON response examples. A single header note SHALL state that all commands accept `--json` and output structured JSON. Agents SHALL discover command-specific parameters via `wallet <command> --help`.

#### Scenario: Agent reads the command reference

- **WHEN** an AI agent reads `skill/COMMANDS.md`
- **THEN** the agent SHALL see only command signatures grouped by domain (e.g., `wallet add expense <amount> <description>`)
- **AND** a header note SHALL state that `--json` enables JSON output for all commands
- **AND** no parameter tables, required/optional markings, or JSON response examples SHALL be present

#### Scenario: Agent discovers parameters at runtime

- **WHEN** an AI agent needs to know flags for a specific wallet command
- **THEN** the agent SHALL run `wallet <command> --help` to discover available options

#### Scenario: Agent needs to know response format

- **WHEN** an AI agent needs to know the JSON response structure for a wallet command
- **THEN** the agent SHALL invoke the command with `--json` to observe the actual output
