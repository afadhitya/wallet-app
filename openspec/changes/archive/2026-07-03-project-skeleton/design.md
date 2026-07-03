## Context

The approved wallet data model defines the SQLite schema and seed data, but the repository needs a concrete Go application structure before feature work can proceed. This change creates the project foundation: a module root, a single `wallet` binary, internal packages for database/query/service/CLI concerns, public configuration loading, sqlc query generation configuration, build/test tooling, and CI wiring.

The skeleton must stay simple enough for an MVP CLI application while preserving clear extension points for upcoming CRUD, budgeting, reporting, forecasting, and Hermes JSON-output workflows.

## Goals / Non-Goals

**Goals:**

- Establish module path `github.com/afadhitya/wallet-app` and a compilable `cmd/wallet` binary.
- Use Cobra for the command tree foundation and keep command implementation isolated under `internal/cli`.
- Use pure-Go SQLite via `modernc.org/sqlite` so local development and cross-compilation do not require CGO.
- Use TOML configuration loaded from XDG-style defaults, with database path, display, AI, and default-account settings.
- Use embedded SQL migrations for the approved data model and provide a migration runner with version tracking.
- Add sqlc configuration and SQL query input directories so repository code can be generated from hand-written SQL.
- Add Makefile and GitHub Actions quality workflow support for build, test, coverage, dependency, and cleanup tasks.

**Non-Goals:**

- Implement full transaction, category, tag, budget, bill, report, or forecast business behavior.
- Build a TUI or long-running service.
- Add external migration frameworks, ORMs, or release automation.
- Implement Hermes integration beyond CLI JSON-output conventions and future-compatible structure.

## Decisions

- Use `cmd/wallet/main.go` as the only binary entry point.
  - Rationale: keeps install/build behavior predictable and leaves all application logic in importable internal packages.
  - Alternative considered: root-level `main.go`; rejected because it mixes command entry code with project root files.

- Organize application code into `internal/db`, `internal/query`, `internal/gen`, `internal/service`, and `internal/cli`.
  - Rationale: separates storage, SQL source, generated repositories, business rules, and command parsing while preventing unintended external imports.
  - Alternative considered: a flatter `pkg/`-heavy layout; rejected because most implementation details are application-internal.

- Keep `pkg/config` public but minimal.
  - Rationale: configuration types and defaults are stable cross-cutting primitives that may be reused by tools or future integrations, while the rest of the application remains internal.
  - Alternative considered: `internal/config`; acceptable, but less flexible for future integrations that need shared config parsing.

- Use Cobra for CLI construction.
  - Rationale: subcommands, flags, help output, and completion support are useful immediately for the planned command tree.
  - Alternative considered: standard-library flag parsing; rejected because nested commands would require custom infrastructure.

- Use `modernc.org/sqlite` for SQLite access.
  - Rationale: pure Go avoids CGO and simplifies local setup and cross-compilation.
  - Alternative considered: `github.com/mattn/go-sqlite3`; rejected because CGO adds development and CI friction for this MVP.

- Use embedded SQL migrations plus a schema version table.
  - Rationale: a single-user CLI app does not need an external migration tool yet, and embedded migrations keep installs self-contained.
  - Alternative considered: external migration binaries or libraries; rejected as unnecessary early complexity.

- Use sqlc for generated repository code.
  - Rationale: it preserves direct SQL control while generating typed Go query methods and models without ORM behavior.
  - Alternative considered: hand-written scan code; rejected because it is repetitive and easier to break as query count grows.

- Provide Makefile targets as the local quality contract.
  - Rationale: contributors and CI can share simple, discoverable commands for build/test/coverage/dependency cleanup.
  - Alternative considered: ad hoc documented commands only; rejected because repeatable targets reduce drift.

## Risks / Trade-offs

- Migration runner correctness risk -> keep migration ordering deterministic, add tests around empty-database initialization, and fail fast on migration errors.
- sqlc generated code may not exist until the developer runs generation -> include sqlc configuration, document generation, and add a Makefile target so generated code can be refreshed consistently.
- A strict 100% coverage gate can slow early development -> start with a small skeleton and tests for foundational behavior so the gate remains achievable.
- Public `pkg/config` can become a dumping ground -> keep only config structs, defaults, and loading helpers there; keep application behavior internal.
- XDG path handling varies by platform -> default to `~/.config/wallet/config.toml` and `~/.local/share/wallet/wallet.db` for the target environment, while centralizing path expansion for future changes.
