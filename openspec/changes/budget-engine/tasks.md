## 1. Query Layer

- [x] 1.1 Add `internal/query/budgets.sql` with queries for creating, finding by ID/name/period, listing active/all, updating fields, marking inactive, and finding the most recent prior period.
- [x] 1.2 Add budget target queries for listing, adding, removing, and replacing category and tag links.
- [x] 1.3 Add aggregate spending queries for category-target and tag-target expense totals within a date range, excluding archived transactions.
- [x] 1.4 Regenerate sqlc output in `internal/gen` and confirm generated code compiles.

## 2. Service Layer

- [x] 2.1 Add budget service parameter/result types for set, list, check, edit, remove, targets, and budget status responses.
- [x] 2.2 Implement period validation and boundary calculation for `monthly`, `weekly`, `yearly`, and `one_time` budgets.
- [x] 2.3 Implement `SetBudget` with positive amount validation, target validation, category/tag resolution, same-name-and-period upsert behavior, and target replacement.
- [x] 2.4 Implement `ListBudgets` with active/all filtering and spent/remaining calculation for each returned budget.
- [x] 2.5 Implement `CheckBudgets` and single-budget check with spent, remaining, percent used, and `ok`/`warning`/`over` status calculation.
- [x] 2.6 Implement recurring period auto-generation for monthly, weekly, and yearly budgets by copying the most recent prior period's amount, notification threshold, active state, and targets.
- [x] 2.7 Implement `EditBudget` for amount, name, notification threshold, added/removed categories, and added/removed tags while preserving unspecified fields.
- [x] 2.8 Implement `RemoveBudget` so removed budgets are excluded from default list and check workflows.
- [x] 2.9 Return consistent validation and not-found errors for CLI rendering.

## 3. CLI Commands

- [x] 3.1 Replace the stubbed `wallet budget` command with `set`, `list`, `check`, `edit`, and `rm` subcommands.
- [ ] 3.2 Add flags for `budget set`: repeatable `--category/-c`, repeatable `--tag/-t`, `--period`, `--from`, `--to`, `--notify`, and inherited `--json`.
- [ ] 3.3 Add flags for `budget list`: `--active`, `--all`, and inherited `--json`.
- [ ] 3.4 Add flags for `budget check`: `--budget/-b`, `--all`, and inherited `--json`.
- [ ] 3.5 Add flags for `budget edit`: `--amount`, `--name`, `--notify`, `--add-category`, `--remove-category`, `--add-tag`, and `--remove-tag`.
- [ ] 3.6 Render stable text output for set, list, check, edit, and remove success/error paths.
- [ ] 3.7 Render stable JSON output for set, list, check, and edit paths.

## 4. Tests

- [ ] 4.1 Add service tests for budget set validation, category targets, tag targets, mixed targets, and same-name-and-period upsert behavior.
- [ ] 4.2 Add service tests for period calculation across monthly, weekly, yearly, one-time, and explicit date inputs.
- [ ] 4.3 Add service tests for list and check spent/remaining/status calculations.
- [ ] 4.4 Add service tests proving non-expense and archived transactions are excluded from spending.
- [ ] 4.5 Add service tests proving mixed category/tag overlap can be double-counted.
- [ ] 4.6 Add service tests for recurring period auto-generation, one-time exclusion, and duplicate current-period prevention.
- [ ] 4.7 Add service tests for budget edit and remove behavior, including missing-budget errors.
- [ ] 4.8 Add CLI integration tests for budget set/list/check/edit/rm text output and database side effects.
- [ ] 4.9 Add CLI integration tests for JSON output and validation failures.

## 5. Verification

- [ ] 5.1 Run sqlc generation verification if available, or regenerate and inspect generated diffs.
- [ ] 5.2 Run `go test ./...` and fix failures.
- [ ] 5.3 Run the repository coverage command and restore the required coverage gate.
- [ ] 5.4 Run `openspec status --change "budget-engine"` and confirm the change is apply-ready.
