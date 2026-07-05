## MODIFIED Requirements

### Requirement: Project README
The project SHALL have a README.md at the repo root that describes the project, its features, installation steps, quick start examples, configuration reference, and a pointer to CLI reference documentation.

#### Scenario: User wants to know available commands
- **WHEN** a user reads the README
- **THEN** the README SHALL direct users to run `wallet --help` or view the auto-generated docs at `docs/cli/` instead of listing all commands in a table

### Requirement: Contributing Guide
The project SHALL have a CONTRIBUTING.md at the repo root covering development setup, project structure, coding conventions, PR process, code generation instructions, and CLI documentation generation instructions.

#### Scenario: Contributor adds or modifies a CLI command
- **WHEN** a contributor reads CONTRIBUTING.md after adding or modifying a Cobra command
- **THEN** they SHALL find instructions to run `make docs` to regenerate the CLI reference documentation in `docs/cli/`

#### Scenario: Contributor wants to understand code generation workflows
- **WHEN** a contributor reads CONTRIBUTING.md
- **THEN** they SHALL find both `make sqlc-gen` (for SQL query changes) and `make docs` (for CLI command changes) as required code generation steps

### Requirement: AGENTS.md references CONTRIBUTING.md for shared concerns
The `AGENTS.md` file SHALL reference `CONTRIBUTING.md` for build commands, project structure, coding conventions, commit conventions, and code generation workflows instead of duplicating them.

#### Scenario: AI agent needs build commands
- **WHEN** an AI agent reads AGENTS.md looking for build commands
- **THEN** the agent SHALL find a reference to CONTRIBUTING.md for the complete build/test/lint commands table

#### Scenario: AI agent needs commit conventions
- **WHEN** an AI agent reads AGENTS.md looking for commit conventions
- **THEN** the agent SHALL find a reference to CONTRIBUTING.md instead of a duplicate list

#### Scenario: make docs added to CONTRIBUTING.md
- **WHEN** a new code generation command (like `make docs`) is added to CONTRIBUTING.md
- **THEN** AGENTS.md SHALL NOT need a separate update — the reference to CONTRIBUTING.md ensures AI agents discover the new command automatically
