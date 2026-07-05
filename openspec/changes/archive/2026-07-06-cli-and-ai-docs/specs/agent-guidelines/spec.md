## ADDED Requirements

### Requirement: SKILL.md delegates detailed reference to split files
The `skill/SKILL.md` file SHALL delegate command reference details (flags, JSON output format) to `COMMANDS.md`, error code details to `ERRORS.md`, and workflow examples to `EXAMPLES.md`, keeping SKILL.md focused on principles, rules, and intent mapping.

#### Scenario: Agent reads SKILL.md command reference section
- **WHEN** an AI agent reaches the section previously containing command reference tables
- **THEN** SKILL.md SHALL contain a reference to `COMMANDS.md` for the detailed command reference

#### Scenario: Agent reads SKILL.md error codes section
- **WHEN** an AI agent reaches the section previously containing the error codes table
- **THEN** SKILL.md SHALL contain a reference to `ERRORS.md` for the full error code reference

## MODIFIED Requirements

### Requirement: Skill installation documented in README
The `README.md` SHALL include instructions for installing the entire `skill/` directory (all files: `SKILL.md`, `COMMANDS.md`, `ERRORS.md`, `EXAMPLES.md`) with AI agentic tools so the skill is registered and auto-loaded during relevant sessions.

#### Scenario: User wants to install the skill for Hermes Agent or OpenClaw
- **WHEN** a user reads the README's installation section
- **THEN** the user SHALL find a sub-section explaining how to copy the entire `skill/` directory (not just `SKILL.md`) to the appropriate agent skills directory for Hermes Agent and OpenClaw
