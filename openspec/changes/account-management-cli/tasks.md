## 1. Service Layer

- [x] 1.1 Add `ListAllAccounts()` method to `internal/service/account.go` wrapping `s.q.ListAllAccounts`
- [x] 1.2 Add `ListAllAccounts` unit test to `internal/service/service_test.go`

## 2. CLI Commands

- [x] 2.1 Create `internal/cli/account.go` with `newAccountCmd()` parent command registering `add`, `list`, `edit`, `archive` subcommands
- [x] 2.2 Implement `runAccountAdd` — parse args and flags, call `svc.CreateAccount`, handle duplicate name, invalid type, and empty name errors, render text/JSON success
- [x] 2.3 Implement `runAccountList` — call `svc.ListAccounts` (or `svc.ListAllAccounts` with `--all`), render table with ID/Name/Type/Currency/Balance/Status columns and total row, render JSON with `printSuccessJSON`
- [x] 2.4 Implement `runAccountEdit` — parse ID and flags (`--name`, `--type`, `--sort-order`), call `svc.UpdateAccount`, handle not-found and empty-name errors, render text/JSON success
- [x] 2.5 Implement `runAccountArchive` — parse ID, check balance, warn if non-zero, prompt confirmation (skip with `--force`), call `svc.ArchiveAccount`, handle not-found error, render text/JSON success

## 3. Root Command Wiring

- [x] 3.1 Add `cmd.AddCommand(newAccountCmd())` to `NewRootCmd()` in `internal/cli/root.go`

## 4. Tests

- [x] 4.1 Create `internal/cli/account_test.go` with unit tests for all four commands covering:
  - Success paths (text output)
  - Success paths (JSON output)
  - Error paths (not found, duplicate name, invalid type, empty name)
  - `--all` flag behavior (list includes archived accounts)
  - `--force` flag behavior (archive skips confirmation)
  - Balance warning on archive
  - Interactive confirmation accept/decline
  - Invalid ID parsing
- [x] 4.2 Create `internal/cli/account_integration_test.go` with integration tests using real SQLite in-memory database
- [x] 4.3 Identify any untestable DB/infrastructure/OS error branches from `account.go` and add them to `coverignore.txt` following the existing entry format

## 5. Documentation

- [x] 5.1 Update README.md to document `wallet account` commands and their flags

## 6. Verification

- [x] 6.1 Run `make fmt` to format source files
- [x] 6.2 Run `make lint` (standard linters per `.golangci.yml`) and fix all issues
- [x] 6.3 Run `make test` to ensure all tests pass
- [x] 6.4 Run `make coverage-check` to ensure 100% coverage
- [x] 6.5 For any genuinely untestable branches (DB/infrastructure/OS errors), add documented exclusions to `coverignore.txt` in the same format as existing entries
