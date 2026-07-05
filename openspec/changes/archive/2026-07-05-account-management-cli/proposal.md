## Why

The service layer already supports account CRUD operations (`CreateAccount`, `UpdateAccount`, `ArchiveAccount`, `ListAccounts`), but no CLI commands expose them to users. Users currently have no way to add, list, edit, or archive accounts through the wallet CLI. The `wallet add` commands accept `--account` flags referencing account names, yet there is no command to manage those accounts themselves. Adding account management CLI commands completes the account lifecycle and aligns with the UX patterns already established by sibling commands (`category`, `tag`, `budget`, `rate`).

## What Changes

- Add `wallet account` parent command with four subcommands: `add`, `list`, `edit`, `archive`.
- `wallet account add <name>` creates a new account with `--type` and `--currency` flags (both optional, defaults: `checking`, `IDR`).
- `wallet account list` displays active accounts with balances in a table, plus `--all` flag to include archived accounts and `--json` for machine output.
- `wallet account edit <id>` updates name, type, and/or sort-order with `--name`, `--type`, `--sort-order` flags. Currency cannot be changed.
- `wallet account archive <id>` soft-deletes an account with `--force` to skip confirmation and `--json` for structured output. Warns when the account has a non-zero balance.
- Wire the new `newAccountCmd()` into `NewRootCmd()` in `root.go`.
- **BREAKING**: None. All additions are new commands.

## Capabilities

### New Capabilities
- `account-management-cli`: CLI commands for full account lifecycle management — create, list, edit, archive — with human-readable tables, JSON output support, and standard validation/error codes.

### Modified Capabilities
- `core-crud`: Adds account management commands (`wallet account add`, `wallet account list`, `wallet account edit`, `wallet account archive`) to the set of core CRUD CLI commands. The backing service-layer and persistence-layer capabilities (account service methods, SQL queries) already exist in Phase 03; this change only introduces the CLI presentation layer.
- `ai-native-cli`: The new commands follow existing JSON output patterns (`success` envelope, `error` envelope with machine-readable codes) established by the AI-native CLI spec.

## Impact

- **New file**: `internal/cli/account.go` — command definitions and run functions for `wallet account add|list|edit|archive`.
- **New file**: `internal/cli/account_test.go` — unit tests with mocked service and integration tests.
- **Modified file**: `internal/cli/root.go` — add `cmd.AddCommand(newAccountCmd())`.
- **Service layer**: No changes needed. `service.Service` already exposes `CreateAccount`, `ListAccounts`, `GetAccountByID`, `UpdateAccount`, `ArchiveAccount`, and `GetAccountBalance`.
- **Database queries**: No changes needed. `internal/query/accounts.sql` already covers `ListAccounts`, `ListAllAccounts`, `CreateAccount`, `UpdateAccount`, `ArchiveAccount`, `GetAccountByID`, `GetAccountByName`, `GetAccountBalance`.
