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
- [ ] 2.7 Implement transaction editing for amount, category, account, date, description, notes, and tag add/remove while recalculating all affected accounts.
- [ ] 2.8 Implement transaction removal as archive/soft-delete with confirmation support at the CLI boundary and affected balance recalculation.
- [ ] 2.9 Implement balance adjustment to set an account to an exact target balance, record an adjustment transaction, and expose adjustments through transaction listing.

## 3. CLI Commands

- [ ] 3.1 Wire `wallet init` to create config directories/files, open the configured database, run migrations, and report idempotent initialization results.
- [ ] 3.2 Replace the placeholder `wallet add` command with `expense`, `income`, and `transfer` subcommands and their flags.
- [ ] 3.3 Implement `wallet list` flags, default current-month behavior, text output, JSON output, and stable validation errors.
- [ ] 3.4 Add `wallet edit <id>` flags for editable transaction fields and tag changes.
- [ ] 3.5 Add `wallet rm <id>` with confirmation prompt, `--force`, and JSON output support.
- [ ] 3.6 Add `wallet category list`, `wallet category add`, `wallet category edit <id>`, and category removal command behavior.
- [ ] 3.7 Add `wallet tag list`, `wallet tag add`, and `wallet tag rm <name>` command behavior.
- [ ] 3.8 Add `wallet adjust <account> <amount> <description>` with notes and JSON output support.

## 4. Output And Errors

- [ ] 4.1 Implement shared amount/date formatting and parsing helpers for IDR minor-unit display and accepted date aliases.
- [ ] 4.2 Implement shared JSON rendering for successful command results and errors.
- [ ] 4.3 Normalize validation errors for missing category, missing account, invalid amount, invalid date, missing tag, insufficient transfer balance warning, and missing transaction.
- [ ] 4.4 Add category-name suggestion support for category-not-found errors.

## 5. Tests And Verification

- [ ] 5.1 Add service unit tests for expense, income, transfer, edit, removal, adjustment, and balance recalculation behavior.
- [ ] 5.2 Add service unit tests for category CRUD and tag CRUD behavior, including duplicate and missing-reference failures.
- [ ] 5.3 Add CLI integration tests for `wallet init`, transaction add/list/edit/rm, category commands, tag commands, adjustment, JSON output, and representative validation failures.
- [ ] 5.4 Run `go fmt` on changed Go files.
- [ ] 5.5 Run sqlc generation and confirm no generated files are stale.
- [ ] 5.6 Run `go test ./...` and the repository lint/coverage commands required by project quality gates.
