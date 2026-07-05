## Context

The service layer (`internal/service/account.go`) already implements `CreateAccount`, `UpdateAccount`, `ArchiveAccount`, `ListAccounts`, `GetAccountByID`, and `GetAccountByName`. The sqlc-generated `Querier` interface (`internal/gen/`) exposes all needed queries including `ListAllAccounts` (includes archived). The CLI framework (`cobra`) already has `withService()` wiring, JSON output helpers (`printSuccessJSON`, `printErrorJSON`), error classification (`classifyError`), and amount formatting (`formatAmount`). The `category.go`, `tag.go`, and `rate.go` command files serve as established patterns for entity management commands.

What's missing: the `account` parent command and its four subcommands in `internal/cli/`, plus registration in `root.go`.

## Goals / Non-Goals

**Goals:**
- Provide `wallet account add <name>` with `--type` and `--currency` flags.
- Provide `wallet account list` with an `--all` flag for archived accounts and `--json` support.
- Provide `wallet account edit <id>` with `--name`, `--type`, and `--sort-order` flags.
- Provide `wallet account archive <id>` with `--force` to skip confirmation and `--json` support.
- Follow existing CLI patterns (command structure, error handling, JSON output).
- 100% test coverage for the new file (unit + integration tests).

**Non-Goals:**
- No changes to the service layer or database schema.
- No new SQL queries. All required queries already exist in `internal/gen/`.
- No currency change support in edit (blocked by design to prevent balance corruption).
- No account deletion (only soft-delete via archive).

## Decisions

1. **One file per entity**: Follow `category.go` pattern — all account command definitions and run functions in `internal/cli/account.go`. This keeps entity management cohesive and matches the existing convention.

2. **`--all` flag wiring**: The `gen.Querier` already has `ListAllAccounts` (includes archived). Add a minimal one-line service method `ListAllAccounts()` to keep the service layer as the single integration point. The CLI list command chooses `ListAccounts` or `ListAllAccounts` based on the `--all` flag.

3. **Archive confirmation prompt**: Match the pattern from `wallet rm` — default prompts for confirmation, `--force` skips. Before prompting, check account balance: if non-zero, display a warning with the balance amount. Use `fmt.Fscanln` on `os.Stdin` for interactive confirmation (consistent with `rm.go`).

4. **Edit: currency is immutable**: The brainstorm explicitly says "Cannot change currency (would break balance)." The `--currency` flag is NOT offered on edit. Only `--name`, `--type`, and `--sort-order` are editable.

5. **Error codes**: Reuse existing error codes. `ACCOUNT_NOT_FOUND` already exists in `classifyError()`. `ErrCodeValidation` and `ErrCodeInvalidInput` cover validation failures. No new error codes needed.

6. **Table formatting for list**: Use the existing `formatAmount()` helper for balance display. Follow `category.go`'s `fmt.Fprintf` approach for column alignment. The table includes ID, Name, Type, Currency, Balance, and Status (active/archived), with a total row at the bottom.

## Testing Strategy

- **Linting**: `.golangci.yml` uses `default: standard` linters. `make lint` covers all packages.
- **Coverage**: Enforced at 100% by `make coverage-check`. The pipeline runs `go test -coverprofile` across all packages (excluding `internal/gen` and `cmd/coverage-filter`), then filters exclusions via `coverignore.txt`.
- **Exclusions**: Infrastructure/OS error branches that cannot be meaningfully tested (DB query errors, stdin read failures, OS-level file errors) are documented in `coverignore.txt` using `go tool cover -func` line ranges. The new account CLI will follow the same exclusion policy — expected untestable branches include:
  - DB query errors from `ListAccounts`, `ListAllAccounts`, `CreateAccount`, `UpdateAccount`, `ArchiveAccount`, `GetAccountByID`
  - `fmt.Fscanln` read failure during archive confirmation prompt (matching `rm.go:67` exclusion)
  - `getService` / `withService` infrastructure errors (config load, DB open, migrate — already covered by existing `init.go` exclusions)

## Risks / Trade-offs

- **Non-zero balance archive**: Archiving an account with a non-zero balance doesn't delete transactions, so historical data is preserved. The warning prompt mitigates accidental archive. → Acceptable; matches the soft-delete design documented in AGENTS.md.
- **Sort-order editing**: The `UpdateAccountParams` in sqlc may not include `sort_order`. The service `UpdateAccount` doesn't expose it. → Add `--sort-order` support to the service method `UpdateAccount` (trivial parameter addition), or defer if not in the current sqlc query. Given the brainstorm lists it, we'll add it.
- **Balance calculation during list**: The `ListAccounts` query returns `balance` from the `accounts` table (pre-calculated). This is accurate and performant. No real-time recalculation needed.

## Open Questions

- None. All decisions are clear from the brainstorming document and existing codebase patterns.
