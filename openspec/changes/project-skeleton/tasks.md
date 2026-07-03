## 1. Module And Tooling

- [x] 1.1 Initialize `go.mod` for module `github.com/afadhitya/wallet-app` with required Cobra, modernc SQLite, TOML, and testify dependencies.
- [x] 1.2 Add `.gitignore` entries for `/bin/`, coverage outputs, and SQLite runtime files.
- [x] 1.3 Add `Makefile` targets for build, run, test, coverage, clean, install, dependency verification, and sqlc generation.
- [x] 1.4 Add `sqlc.yaml` configured for SQLite schema files in `internal/db/migrations`, query files in `internal/query`, and generated Go output in `internal/gen`.

## 2. Project Structure

- [x] 2.1 Create the `cmd/wallet`, `internal/db`, `internal/query`, `internal/gen`, `internal/service`, `internal/cli`, and `pkg/config` package directories.
- [x] 2.2 Add a compilable `cmd/wallet/main.go` that executes the wallet CLI root command.
- [x] 2.3 Add package placeholders or minimal implementations so `go test ./...` discovers and compiles all skeleton packages.

## 3. Configuration

- [ ] 3.1 Implement `pkg/config` types for database, display, AI, and defaults settings.
- [ ] 3.2 Implement default configuration values for wallet database path, currency, date format, first day of week, JSON output behavior, and default account.
- [ ] 3.3 Implement TOML config loading with path expansion for the wallet config file.
- [ ] 3.4 Add unit tests for default configuration and TOML override parsing.

## 4. Database And Migrations

- [x] 4.1 Implement `internal/db.Open` using `modernc.org/sqlite` with WAL mode and foreign keys enabled.
- [x] 4.2 Add embedded migration file `internal/db/migrations/001_initial.sql` containing the approved wallet data model schema and default category seed data.
- [x] 4.3 Implement migration version tracking and ordered migration execution.
- [x] 4.4 Add tests for empty-database migration and idempotent migration re-runs.

## 5. CLI Scaffold

- [ ] 5.1 Implement `internal/cli` root command construction for the `wallet` binary.
- [ ] 5.2 Add initial subcommand scaffolding for `init`, `add`, `list`, `category`, `tag`, `budget`, `bill`, `report`, and `forecast`.
- [ ] 5.3 Add shared `--json` output option support for machine-readable CLI output.
- [ ] 5.4 Add CLI tests verifying root command construction and planned subcommand registration.

## 6. CI And Verification

- [x] 6.1 Add GitHub Actions workflow for pushes to `main` and `brainstorming` and pull requests to `main`.
- [x] 6.2 Configure the workflow to set up Go from `go.mod`, run linting, run tests with atomic coverage, and fail below 100% coverage.
- [x] 6.3 Configure workflow artifacts for `coverage.out` on every run and `coverage.html` on coverage failure.
- [ ] 6.4 Run local verification with `go test ./...`, `go build ./cmd/wallet`, and the relevant Makefile targets.
