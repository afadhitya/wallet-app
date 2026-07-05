## ADDED Requirements

### Requirement: CLI reference documentation is auto-generated
The system SHALL auto-generate Markdown CLI reference documentation from the Cobra command tree using the `github.com/spf13/cobra/doc` package.

#### Scenario: All visible commands have generated docs
- **WHEN** the documentation generation command runs
- **THEN** one Markdown file per visible command is generated, with filenames matching the command hierarchy (e.g., `wallet_add_expense.md`)

#### Scenario: Hidden commands are excluded
- **WHEN** the documentation generation command runs
- **THEN** hidden commands (including the `docs` command itself) SHALL NOT produce generated Markdown files

#### Scenario: Generated docs include flags and examples
- **WHEN** a command's documentation is generated
- **THEN** the Markdown file SHALL include the command's usage line, short and long descriptions, all flags with types and defaults, and example text where configured

### Requirement: Makefile target for documentation generation
The project SHALL provide a `make docs` target that generates CLI reference documentation to the `docs/cli/` directory.

#### Scenario: Developer runs make docs
- **WHEN** a developer runs `make docs`
- **THEN** the wallet CLI is invoked to generate Markdown files into `docs/cli/` and a success message is printed

### Requirement: Generated docs are git-ignored
The `docs/cli/` directory SHALL be listed in `.gitignore` so generated documentation is not committed to the repository.

#### Scenario: Git status after generation
- **WHEN** documentation is generated and `git status` is run
- **THEN** the `docs/cli/` directory SHALL NOT appear as untracked or modified files

### Requirement: Docs command is hidden from user-facing help
The `wallet docs` command and its subcommands SHALL be marked as hidden so they do not appear in `wallet --help` or `wallet <cmd> --help` output.

#### Scenario: User runs wallet --help
- **WHEN** a user runs `wallet --help`
- **THEN** the `docs` command SHALL NOT appear in the available commands list

#### Scenario: Developer runs wallet docs markdown directly
- **WHEN** a developer runs `wallet docs markdown`
- **THEN** the command SHALL execute successfully and generate docs even though it is hidden from help
