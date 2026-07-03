## Purpose

TBD - Captures the wallet application project skeleton and entry point structure.

## Requirements

### Requirement: Wallet command entry point
The application SHALL provide a compilable `wallet` command binary from `cmd/wallet` using module path `github.com/afadhitya/wallet-app`.

#### Scenario: Build wallet command
- **WHEN** a developer runs `go build ./cmd/wallet` from the repository root
- **THEN** the wallet command package compiles successfully

#### Scenario: Execute root command
- **WHEN** the compiled wallet command is executed without a subcommand
- **THEN** it initializes the Cobra root command and displays valid command help or root command output without panicking

### Requirement: Application package layout
The repository SHALL separate wallet application concerns into internal packages for database access, SQL query sources, generated query code, services, and CLI commands.

#### Scenario: Source tree contains skeleton packages
- **WHEN** the project skeleton is present
- **THEN** the repository contains `internal/db`, `internal/query`, `internal/gen`, `internal/service`, and `internal/cli` locations for the corresponding application layers

### Requirement: Configuration loading
The application SHALL define TOML-backed configuration with defaults for database path, display settings, AI output behavior, and default account.

#### Scenario: Load default config
- **WHEN** configuration is loaded without an existing user config file
- **THEN** defaults include database path `~/.local/share/wallet/wallet.db`, currency `IDR`, date format `2006-01-02`, first day of week `monday`, JSON output enabled for AI use, and an empty or explicit default account value

#### Scenario: Load TOML config file
- **WHEN** a TOML config file exists at the configured wallet config path
- **THEN** the application parses database, display, AI, and defaults sections into typed configuration values

### Requirement: SQLite connection
The application SHALL open wallet SQLite databases through a shared database package using the pure-Go SQLite driver.

#### Scenario: Open SQLite database
- **WHEN** the database package opens a wallet database path
- **THEN** it uses the `modernc.org/sqlite` driver and enables WAL journaling and foreign key enforcement for application connections

### Requirement: Embedded migrations
The application SHALL embed SQL migration files and provide a migration runner that applies schema changes in order with version tracking.

#### Scenario: Migrate empty database
- **WHEN** the migration runner executes against an empty wallet database
- **THEN** it creates the schema version tracking table and applies the initial wallet schema migration

#### Scenario: Re-run migrations
- **WHEN** the migration runner executes after all available migrations have already been applied
- **THEN** it does not reapply completed migrations and exits successfully

### Requirement: sqlc setup
The repository SHALL configure sqlc to generate Go query code from SQL files under `internal/query` using schema files from `internal/db/migrations`.

#### Scenario: sqlc configuration is available
- **WHEN** a developer runs sqlc using the repository configuration
- **THEN** generated Go code is targeted to `internal/gen` with package name `gen`, JSON tags, DB tags, empty slices, pointer result structs, and a Querier interface enabled

### Requirement: CLI command scaffold
The CLI layer SHALL define a root wallet command and placeholders for the planned command tree.

#### Scenario: Root command exposes planned structure
- **WHEN** a developer inspects CLI command registration
- **THEN** it includes the root command plus initial subcommand scaffolding for init, add, list, category, tag, budget, bill, report, and forecast workflows

### Requirement: JSON output convention
Wallet CLI commands SHALL reserve a `--json` output mode so future automation can consume structured output.

#### Scenario: Command supports JSON flag convention
- **WHEN** a wallet command is constructed
- **THEN** the CLI layer provides a consistent `--json` flag or shared output option for commands intended to produce machine-readable results
