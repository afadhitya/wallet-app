## Why

Four files in `internal/service/` exceed 500 lines (transaction.go at 753, planned_payment.go at 691, budget.go at 662, report.go at 503), making them hard to navigate, test, and maintain. This refactor extracts each bloated domain into its own sub-package with a shared cross-cutting package to keep the codebase scalable.

## What Changes

- Extract `internal/service/transaction.go` into `internal/service/transaction/` sub-package with files per concern (expense, income, transfer, adjustment, edit, list)
- Extract `internal/service/planned_payment.go` into `internal/service/plannedpayment/` sub-package (create/edit, pay/skip, states, rrule parser)
- Extract `internal/service/budget.go` into `internal/service/budget/` sub-package (create, check, spending, edit)
- Extract `internal/service/report.go` into `internal/service/report/` sub-package (breakdown, export)
- Introduce `internal/service/shared/` sub-package for cross-cutting types and helpers (error types, resolvers, currency helpers, date parsing)
- Embed sub-package managers into `*Service` via struct embedding — zero API changes for callers
- Move tests for extracted methods to their respective sub-package `_test.go` files

## Capabilities

### New Capabilities

None. This is a pure internal refactoring — no new user-facing capabilities.

### Modified Capabilities

None. No requirement changes to existing capabilities. All CLI commands, business rules, and outputs remain identical.

## Impact

- All files under `internal/service/` (new sub-package structure, `service.go` updated with manager embedding)
- `internal/cli/helpers.go` — no changes needed (embedded methods promote to `*Service`)
- Error type imports for CLI and test code update from `service.NotFoundError` to `shared.NotFoundError`
- Test files move from `service_test.go` to sub-package `_test.go` files
- Build and test scripts unaffected (same `go build`/`go test` commands)
