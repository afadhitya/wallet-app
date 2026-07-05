# Agent Guidelines

## Purpose

Guidelines and documentation for AI agents working on this codebase. TBD.

## Requirements

### Requirement: CLI flag-argument separator documented

The `skill/SKILL.md` SHALL document that flags MUST appear before the `--` separator, and all arguments after `--` are treated as positional arguments. The documentation MUST include a visual example showing the correct command structure.

#### Scenario: Agent reads CLI separator documentation

- **WHEN** an AI agent reads `skill/SKILL.md`
- **THEN** the agent SHALL see documentation explaining that `wallet adjust "Bunga Bank" --json -- -3612 "Initial balance"` is the correct syntax, with flags (`--json`) before `--` and positional args (`-3612`, `"Initial balance"`) after it

### Requirement: Database access restricted to wallet CLI

The `skill/SKILL.md` SHALL state that AI agents MUST NOT access the SQLite database directly or create scripts that manipulate database data. Agents MUST use the `wallet` CLI for all data operations.

#### Scenario: Agent needs to insert a transaction

- **WHEN** an AI agent needs to create a new transaction in the database
- **THEN** the agent SHALL use a `wallet add` CLI command instead of writing SQL or running a script that opens the database file

#### Scenario: Agent needs to query data

- **WHEN** an AI agent needs to read data from the database
- **THEN** the agent SHALL use a `wallet` CLI command (e.g., `wallet list`, `wallet show`) instead of executing raw SQL queries

### Requirement: SKILL.md delegates detailed reference to split files

The `skill/SKILL.md` file SHALL delegate command reference details (flags, JSON output format) to `COMMANDS.md`, error code details to `ERRORS.md`, and workflow examples to `EXAMPLES.md`, keeping SKILL.md focused on principles, rules, and intent mapping.

#### Scenario: Agent reads SKILL.md command reference section

- **WHEN** an AI agent reaches the section previously containing command reference tables
- **THEN** SKILL.md SHALL contain a reference to `COMMANDS.md` for the detailed command reference

#### Scenario: Agent reads SKILL.md error codes section

- **WHEN** an AI agent reaches the section previously containing the error codes table
- **THEN** SKILL.md SHALL contain a reference to `ERRORS.md` for the full error code reference

### Requirement: Skill installation documented in README

The `README.md` SHALL include instructions for installing the entire `skill/` directory (all files: `SKILL.md`, `COMMANDS.md`, `ERRORS.md`, `EXAMPLES.md`) with AI agentic tools so the skill is registered and auto-loaded during relevant sessions.

#### Scenario: User wants to install the skill for Hermes Agent or OpenClaw

- **WHEN** a user reads the README's installation section
- **THEN** the user SHALL find a sub-section explaining how to copy the entire `skill/` directory (not just `SKILL.md`) to the appropriate agent skills directory for Hermes Agent and OpenClaw
