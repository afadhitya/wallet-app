## Why

The wallet app has an approved data model, but it still needs the runnable Go project foundation that every later CLI, service, database, reporting, and AI-integration phase will build on. Establishing the module, package structure, configuration conventions, SQLite integration, build tooling, and CI now reduces rework and gives future changes a stable implementation target.

## What Changes

- Initialize the repository as module `github.com/afadhitya/wallet-app` with a `wallet` command entry point.
- Add the initial package layout for database access, sqlc query inputs/generated code, business services, Cobra CLI commands, and configuration loading.
- Introduce SQLite database opening and embedded migration infrastructure using the approved wallet data model schema.
- Add TOML configuration support with XDG-style default config and data paths.
- Add sqlc configuration for type-safe repository generation from SQL files.
- Add local build, test, coverage, dependency, and cleanup commands through a Makefile.
- Add GitHub Actions coverage workflow expectations for pushes and pull requests.
- Add ignore rules for build outputs, coverage artifacts, and SQLite database files.

## Capabilities

### New Capabilities
- `wallet-project-skeleton`: Defines the runnable Go wallet application skeleton, package layout, CLI foundation, configuration defaults, SQLite integration, migration scaffolding, and sqlc setup.

### Modified Capabilities
- `go-project-quality`: Extends project quality expectations to include Makefile targets, sqlc/code-generation support, repository ignore rules, and CI coverage artifact behavior required by the wallet project skeleton.
- `wallet-data-model`: Connects database initialization requirements to the embedded migration file and startup migration runner introduced by the project skeleton.

## Impact

- Affects repository structure under `cmd/`, `internal/`, and `pkg/`.
- Adds Go dependencies for Cobra, pure-Go SQLite, TOML config parsing, and test assertions.
- Adds sqlc development configuration and query input locations.
- Adds build tooling through `Makefile`, `.gitignore`, and GitHub Actions workflow files.
- Provides the foundation for later CRUD, reporting, forecasting, and Hermes-oriented JSON CLI output work.
