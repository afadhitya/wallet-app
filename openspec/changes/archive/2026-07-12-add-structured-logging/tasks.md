## 1. CLI Infrastructure

- [x] 1.1 Add `-v`/`--verbose` count flag and `--log-file` string flag to root Cobra command in `internal/cli/root.go`
- [x] 1.2 Create `internal/cli/logging.go` with `MultiHandler` implementation wrapping two `slog.Handler` instances
- [x] 1.3 Add `newLogger(cmd *cobra.Command) *slog.Logger` to `internal/cli/helpers.go` that constructs the logger with correct level and optional dual output
- [x] 1.4 Update `getService()` and `withService()` signatures in `internal/cli/helpers.go` to create and pass a `*slog.Logger` to `db.Open()`, `db.Migrate()`, and `service.New()`
- [x] 1.5 Add `safeGetService()` helper in `internal/cli/helpers.go` if needed for error-resilient service retrieval

## 2. Database Layer

- [x] 2.1 Add `*slog.Logger` parameter to `db.Open()` in `internal/db/db.go`, with DEBUG log on opening and INFO log with `journal_mode` on success
- [x] 2.2 Add `*slog.Logger` parameter to `db.Migrate()` in `internal/db/db.go`, with INFO log per applied migration and DEBUG log per skipped migration
- [x] 2.3 Update `internal/testdb/testdb.go` `Open()` to accept and pass a `*slog.Logger` parameter

## 3. Service Layer

- [x] 3.1 Add `logger *slog.Logger` field to `Service` struct in `internal/service/service.go`
- [x] 3.2 Update `New()` and `NewWithQuerier()` constructors in `internal/service/service.go` to accept and store `*slog.Logger`
- [x] 3.3 Add entry/exit/error log calls to all public methods in `internal/service/account.go`
- [x] 3.4 Add entry/exit/error log calls to all public methods in `internal/service/budget.go`
- [x] 3.5 Add entry/exit/error log calls to all public methods in `internal/service/category.go`
- [x] 3.6 Add entry/exit/error log calls to all public methods in `internal/service/currency.go`
- [x] 3.7 Add entry/exit/error log calls to all public methods in `internal/service/forecast.go`
- [x] 3.8 Add entry/exit/error log calls to all public methods in `internal/service/planned_payment.go`
- [x] 3.9 Add entry/exit/error log calls to all public methods in `internal/service/report.go`
- [x] 3.10 Add entry/exit/error log calls to all public methods in `internal/service/tag.go`
- [x] 3.11 Add entry/exit/error log calls to all public methods in `internal/service/transaction.go`

## 4. Test Updates

- [x] 4.1 Update `service.NewWithQuerier()` calls in all `internal/service/*_test.go` files to pass a discard logger
- [x] 4.2 Update `testdb.Open()` calls in all `internal/service/*_test.go` files to pass a discard logger
- [x] 4.3 Update `db.Open()` and `db.Migrate()` calls in `internal/cli/*_test.go` files to pass a discard logger
- [x] 4.4 Update any other test files calling `service.New()`, `db.Open()`, or `db.Migrate()` to pass a logger

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify compilation
- [x] 5.2 Run `go test ./...` to verify all tests pass
- [x] 5.3 Run `make coverage-check` to verify 100% coverage is maintained
- [x] 5.4 Run `make lint` to verify style conformance
