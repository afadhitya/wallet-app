## Why

The wallet app currently has no logging, making it impossible to debug issues or understand application behavior without modifying and recompiling code. Adding structured logging with `log/slog` provides observability across all layers — CLI, service, and database — without introducing external dependencies.

## What Changes

- Add `-v`/`--verbose` and `--log-file` global persistent flags to the root Cobra command for controlling log level and output destination
- Inject `*slog.Logger` into the `Service` struct, propagating it from CLI through service to database layers
- Add a `MultiHandler` to support simultaneous human-readable stderr output and JSON file output
- Add structured log calls at entry, exit, and error paths in all service layer public methods
- Add structured log calls in `db.Open()` and `db.Migrate()` for database lifecycle events
- Update all test files to pass a discard logger to `service.New()`/`NewWithQuerier()` and `db.Open()`
- Update `open.go` to accept a `*slog.Logger` parameter — **BREAKING** signature change for `db.Open()`, `db.Migrate()`, and `service.New()`/`NewWithQuerier()`

## Capabilities

### New Capabilities
- `structured-logging`: Structured logging across CLI, service, and database layers using Go's `log/slog` with configurable verbosity levels and dual output (stderr + JSON file)

### Modified Capabilities
<!-- No existing spec-level behavior changes -->

## Impact

- Affected code: `internal/cli/root.go`, `internal/cli/helpers.go`, `internal/service/service.go`, all `internal/service/*.go` files, `internal/db/db.go`, `internal/testdb/testdb.go`, all `*_test.go` files
- New file: `internal/cli/logging.go` (MultiHandler)
- No new dependencies (uses `log/slog` from Go stdlib)
- No changes to `internal/gen/` (sqlc-generated code untouched)
- No changes to CLI `--json` output behavior
