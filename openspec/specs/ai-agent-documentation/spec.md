# AI Agent Documentation

## Purpose

Documentation files for AI agents working with the wallet CLI. TBD.

## Requirements

### Requirement: AI agent documentation is split into focused files

The `skill/` directory SHALL contain separate documentation files for different concerns: `COMMANDS.md` for command reference, `ERRORS.md` for error codes and recovery, and `EXAMPLES.md` for common workflows.

#### Scenario: Agent needs command reference

- **WHEN** an AI agent needs to look up a specific command's flags, arguments, and JSON output format
- **THEN** the agent SHALL find the information in `skill/COMMANDS.md`

#### Scenario: Agent encounters an error

- **WHEN** an AI agent encounters an error response from the wallet CLI
- **THEN** the agent SHALL find the error code, meaning, and recovery actions in `skill/ERRORS.md`

#### Scenario: Agent needs workflow guidance

- **WHEN** an AI agent needs to know how to accomplish a multi-step task (e.g., "track subscriptions", "payday routine")
- **THEN** the agent SHALL find ready-to-use command sequences in `skill/EXAMPLES.md`

### Requirement: COMMANDS.md groups commands by domain

The `skill/COMMANDS.md` file SHALL organize commands into domain-grouped sections (Transaction, Account, Category, Tag, Budget, Bill, Forecast, Report, Rate) rather than a flat alphabetical list.

#### Scenario: Agent needs account-related commands

- **WHEN** an AI agent searches for account commands
- **THEN** the agent SHALL find all account subcommands (`account add`, `account list`, `account edit`, `account archive`) grouped under a single "Account Commands" heading

### Requirement: ERRORS.md includes recovery patterns

The `skill/ERRORS.md` file SHALL document each error code with its meaning, cause, and suggested recovery action, plus common recovery patterns as a reference table.

#### Scenario: Agent receives CATEGORY_NOT_FOUND error

- **WHEN** an AI agent receives a `CATEGORY_NOT_FOUND` error
- **THEN** the agent SHALL find the error code in ERRORS.md with the recovery suggestion "List categories, suggest closest match"

### Requirement: SKILL.md references split files

The `skill/SKILL.md` file SHALL be refactored to focus on core principles, rules, and intent mapping, and SHALL reference `COMMANDS.md`, `ERRORS.md`, and `EXAMPLES.md` for detailed reference material.

#### Scenario: Agent reads SKILL.md for detailed command syntax

- **WHEN** an AI agent reads `skill/SKILL.md` looking for detailed command flags and JSON output structure
- **THEN** the agent SHALL be directed to `COMMANDS.md` for the complete command reference

#### Scenario: SKILL.md retains core principles

- **WHEN** an AI agent reads `skill/SKILL.md`
- **THEN** the file SHALL still contain core principles (always use `--json`, parse envelope, never auto-create, present friendly output), the intent mapping table, and behavioral rules

### Requirement: EXAMPLES.md covers common workflows

The `skill/EXAMPLES.md` file SHALL document common multi-step workflows with ready-to-use command sequences.

#### Scenario: Agent helps with payday routine

- **WHEN** an AI agent needs to guide a user through a payday routine
- **THEN** `skill/EXAMPLES.md` SHALL provide a sequence of commands for recording salary, checking budgets, and reviewing upcoming bills

#### Scenario: Agent helps set up subscription tracking

- **WHEN** an AI agent needs to help a user set up subscription tracking
- **THEN** `skill/EXAMPLES.md` SHALL provide example commands for adding recurring bills and checking forecast
