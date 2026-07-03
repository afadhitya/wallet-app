## 1. Database And Query Layer

- [x] 1.1 Add sqlc queries for accounts, categories, tags, transactions, transaction tags, list filters, and balance recalculation.
- [x] 1.2 Add the minimal migration support needed for transaction/category archiving and update timestamps if the current schema cannot support required soft-delete behavior.
- [x] 1.3 Regenerate `internal/gen` from `sqlc.yaml` and verify generated code compiles.
- [x] 1.4 Add database test helpers that create isolated migrated SQLite databases with seed data for service and CLI tests.

## 2. Service Layer

- [x] 2.1 Implement account lookup, listing, creation/update/archive helpers, and exact ID-or-name resolution.
- [x] 2.2 Implement category listing, creation, editing, active-only lookup, and removal behavior that preserves historical transaction references.
- [x] 2.3 Implement tag listing, creation, deletion, and transaction-tag association helpers without auto-creating tags during transaction entry.
- [x] 2.4 Implement transaction create methods for expenses and income with validation, date parsing, category/account/tag resolution, persistence, and balance recalculation.
- [x] 2.5 Implement transfer creation with distinct source/destination validation, insufficient-balance warning reporting, persistence, and source/destination balance recalculation.
- [x] 2.6 Implement transaction listing filters for month, date range, account, category, tag, type, limit, ordering, and totals.
- [x] 2.7 Implement transaction editing for amount, category, account, date, description, notes, and tag add/remove while recalculating all affected accounts.
- [x] 2.8 Implement transaction removal as archive/soft-delete with confirmation support at the CLI boundary and affected balance recalculation.
- [x] 2.9 Implement balance adjustment to set an account to an exact target balance, record an adjustment transaction, and expose adjustments through transaction listing.

## 3. CLI Commands

- [x] 3.1 Wire `wallet init` to create config directories/files, open the configured database, run migrations, and report idempotent initialization results.
- [x] 3.2 Replace the placeholder `wallet add` command with `expense`, `income`, and `transfer` subcommands and their flags.
- [x] 3.3 Implement `wallet list` flags, default current-month behavior, text output, JSON output, and stable validation errors.
- [x] 3.4 Add `wallet edit <id>` flags for editable transaction fields and tag changes.
- [x] 3.5 Add `wallet rm <id>` with confirmation prompt, `--force`, and JSON output support.
- [x] 3.6 Add `wallet category list`, `wallet category add`, `wallet category edit <id>`, and category removal command behavior.
- [x] 3.7 Add `wallet tag list`, `wallet tag add`, and `wallet tag rm <name>` command behavior.
- [x] 3.8 Add `wallet adjust <account> <amount> <description>` with notes and JSON output support.

## 4. Output And Errors

- [x] 4.1 Implement shared amount/date formatting and parsing helpers for IDR minor-unit display and accepted date aliases.
- [x] 4.2 Implement shared JSON rendering for successful command results and errors.
- [x] 4.3 Normalize validation errors for missing category, missing account, invalid amount, invalid date, missing tag, insufficient transfer balance warning, and missing transaction.
- [x] 4.4 Add category-name suggestion support for category-not-found errors.

## 5. Tests And Verification

- [x] 5.1 Add service unit tests for expense, income, transfer, edit, removal, adjustment, and balance recalculation behavior.
- [x] 5.2 Add service unit tests for category CRUD and tag CRUD behavior, including duplicate and missing-reference failures.
- [x] 5.3 Add CLI integration tests for `wallet init`, transaction add/list/edit/rm, category commands, tag commands, adjustment, JSON output, and representative validation failures.
- [x] 5.4 Run `go fmt` on changed Go files.
- [x] 5.5 Run sqlc generation and confirm no generated files are stale.
- [x] 5.6 Run `go test ./...` and the repository lint/coverage commands required by project quality gates.
- [x] 5.7 Pull the GitHub Actions coverage output or reproduce it locally and identify every file/function/branch keeping total coverage below 100%.
- [x] 5.8 Add focused tests for uncovered service, CLI, database, output formatting, JSON rendering, validation, and error-handling paths reported by the coverage profile.
- [x] 5.9 Remove or simplify unreachable implementation code that cannot be meaningfully exercised instead of excluding it from coverage.
- [x] 5.10 Rerun the exact repository coverage command used by CI and confirm `go tool cover -func=coverage.out` reports total coverage of 100%.
- [x] 5.11 Re-run GitHub Actions or inspect the next workflow run to verify the coverage gate passes for the core CRUD change.

## 6. Coverage Profile Remediation

- [x] 6.1 Convert the pasted atomic coverage profile into an actionable uncovered-block checklist grouped by package after excluding generated `internal/gen` code: `cmd/wallet`, `internal/db`, `internal/cli`, `internal/testdb`, `internal/service`, and `pkg/config`.
- [x] 6.2 Update the Makefile coverage target and GitHub Actions coverage command to build the package list with `go list ./... | grep -v '/internal/gen$'` so sqlc-generated code is excluded from coverage while still being compiled by normal `go test ./...`.
- [x] 6.3 Verify `internal/gen` remains generated, checked for staleness, and compiled by tests, but no longer appears in `coverage.out` or `go tool cover -func=coverage.out`.
- [x] 6.4 Add package-local tests for `internal/testdb/testdb.go` covering successful migrated database setup, returned query helper usability, cleanup behavior, and reachable setup failure paths.
- [x] 6.5 Add CLI tests for uncovered command error branches in `add.go`, `adjust.go`, `category.go`, `edit.go`, `list.go`, `rm.go`, `tag.go`, and `helpers.go`, including invalid IDs, missing args, service errors, JSON rendering, text rendering, and confirmation accept/decline paths.
- [x] 6.6 Add tests for uncovered formatting branches in `internal/cli/format.go`, including zero/negative amounts, thousands grouping, valid aliases, explicit dates, and invalid date input.
- [x] 6.7 Add service tests for uncovered account/category/tag CRUD error paths, including duplicate names, missing rows, archived rows, invalid parent category, empty names, and query failure propagation.
- [x] 6.8 Add service tests for uncovered transaction branches, including missing accounts/categories/tags, invalid amounts/dates, transfer same-account validation, insufficient-balance warning path, edit tag add/remove failures, archive failures, adjustment increase/decrease/no-op, and balance recalculation query errors.
- [x] 6.9 Add tests or refactors for uncovered `cmd/wallet/main.go`, `internal/db/db.go`, `internal/service/service.go`, and `pkg/config/config.go` branches reported by the pasted profile, prioritizing explicit tests for error branches before removing unreachable code.
- [x] 6.10 Run the updated coverage command after each package group and confirm the atomic profile no longer contains zero-count blocks for the touched package or any `internal/gen` entries.
- [x] 6.11 Re-run GitHub Actions or inspect the next workflow run to verify the 100% coverage gate passes with generated code excluded.

## 7. Infrastructure Coverage Exclusions

- [x] 7.1 Identify the exact remaining 0.9% uncovered branches in CLI init, mkdir, rm, and tag infrastructure error handling and record the file/function or coverage block for each exclusion.
- [x] 7.2 Update the coverage tooling to exclude only the documented hard-to-test OS/infrastructure branches, while keeping `internal/gen` excluded and keeping all application packages in the coverage gate.
- [x] 7.3 Ensure the exclusion mechanism is auditable in CI and does not hide business validation, service behavior, JSON/text rendering, normal command errors, or database logic.
- [x] 7.4 Re-run the CI-equivalent coverage command and confirm included coverage is 100% after excluding generated code and the documented OS/infrastructure branches.
- [x] 7.5 Update local developer documentation or Makefile help so contributors understand that generated code and documented OS-level infrastructure failure branches are excluded, but new application logic still requires full coverage.
