# Service Layer Refactor — Extract Bloated Files into Sub-Packages

**Date:** 2026-07-11
**Status:** Spec (not yet implemented)

## Problem

Four files in `internal/service/` exceed 500 lines, making them hard to navigate and reason about:

| File | Lines | Contains |
|------|-------|----------|
| `transaction.go` | 753 | 4 transaction types (expense, income, transfer, adjustment), edit, delete, list with complex filtering |
| `planned_payment.go` | 691 | Full bill lifecycle + ~200 lines of RRULE parser logic |
| `budget.go` | 662 | Budget lifecycle + recurring rollover + spending calculation |
| `report.go` | 503 | Multi-breakdown reports + CSV export |

## Solution

Extract each bloated domain into a sub-package under `internal/service/`. Introduce a `shared/` sub-package for cross-cutting types and functions to avoid circular imports.

### Folder structure after refactor

```
internal/service/
├── service.go                         # Service struct, error type wrappers, constructors
├── account.go                         # unchanged
├── category.go                        # unchanged
├── tag.go                             # unchanged
├── currency.go                        # unchanged
├── forecast.go                        # unchanged
│
├── shared/
│   ├── shared.go                      # NotFoundError, ValidationError, sentinel errors
│   ├── resolvers.go                   # ResolveCategory, ResolveAccount, ResolveTag (take gen.Querier)
│   ├── currency.go                    # GetBaseCurrency, Convert, LoadRates/SaveRates vars
│   └── dates.go                       # ParseDate, ParseMonth (pure functions)
│
├── transaction/
│   ├── manager.go                     # Manager struct + NewManager + recalcBalance
│   ├── types.go                       # All params/result structs (11 types)
│   ├── expense.go                     # AddExpense
│   ├── income.go                      # AddIncome
│   ├── transfer.go                    # AddTransfer + TransferResult
│   ├── adjustment.go                  # AdjustBalance + AdjustBalanceParams/Result
│   ├── edit.go                        # EditTransaction + RemoveTransaction
│   └── list.go                        # ListTransactions + filter helpers + resolveBaseFields
│
├── plannedpayment/
│   ├── manager.go                     # Manager + NewManager + ListPlannedPayments + ListDuePlannedPayments
│   ├── types.go                       # Create/Edit/Pay params + results + ListDueFilter
│   ├── create_edit.go                 # CreatePlannedPayment + EditPlannedPayment
│   ├── pay_skip.go                    # PayPlannedPayment + SkipPlannedPayment
│   ├── states.go                      # PausePlannedPayment, ResumePlannedPayment, DeletePlannedPayment
│   └── rrule.go                       # calc helpers + RRULE parser + validateRRULE
│
├── budget/
│   ├── manager.go                     # Manager + NewManager + ListBudgets
│   ├── types.go                       # All params/results (7 types)
│   ├── create.go                      # SetBudget + updateExistingBudget
│   ├── check.go                       # CheckBudgets + rollover + resolveBudget + buildCheckResult
│   ├── spending.go                    # calculateSpending
│   └── edit.go                        # EditBudget + RemoveBudget
│
└── report/
    ├── manager.go                     # Manager + NewManager + GenerateReport (dispatcher)
    ├── types.go                       # ReportParams, ReportFilters, ReportResult, all *Row types
    ├── breakdown.go                   # generateMonthlySummary, generateCategoryBreakdown, etc.
    └── export.go                      # GenerateExportRows + DefaultExportFilename
```

### Manager Struct Pattern

Every sub-package manager follows the same pattern:

```go
type Manager struct {
    q gen.Querier
}

func NewManager(q gen.Querier) *Manager {
    return &Manager{q: q}
}
```

All methods are on `*Manager`. No manager imports the parent `service` package.

### Service Composition

```go
type Service struct {
    q  gen.Querier
    db *sql.DB

    *transaction.Manager
    *plannedpayment.Manager
    *budget.Manager
    *report.Manager
}
```

Embedding promotes all `Manager` methods to `*Service` — callers (CLI, tests) see no API change.

`New` and `NewWithQuerier` wire all four managers via their constructors.

### The `shared/` Sub-Package

Prevents circular imports: parent `service` imports sub-packages; sub-packages import `shared/`. `shared/` never imports `service/`.

| File | What it provides | Dependencies |
|------|-----------------|-------------|
| `shared.go` | `NotFoundError`, `ValidationError`, `ErrNotFound`, `ErrDuplicateName`, `ErrInvalidAmount`, `ErrMissingField` | stdlib |
| `resolvers.go` | `ResolveCategory(q, name)`, `ResolveAccount(q, name)`, `ResolveTag(q, name)` | `gen.Querier`, stdlib |
| `currency.go` | `GetBaseCurrency()`, `Convert(amount, currency)`, `LoadRates`/`SaveRates` vars | `config.RatesConfig`, stdlib |
| `dates.go` | `ParseDate(input)`, `ParseMonth(input)` — pure functions | stdlib |

Parent `*Service` keeps thin delegation methods (`s.ResolveCategory(...)` calls `shared.ResolveCategory(s.q, ...)`) so non-bloated domain files (forecast, account) don't change their calls.

### Cross-Call Migration

Every `s.SomeMethod(...)` call in the extracted code is replaced:

| Old call (on `*Service`) | Replacement in sub-package |
|---|---|
| `s.ResolveCategory(name)` | `shared.ResolveCategory(m.q, name)` |
| `s.ResolveAccount(name)` | `shared.ResolveAccount(m.q, name)` |
| `s.ResolveTag(name)` | `shared.ResolveTag(m.q, name)` |
| `s.AddTransactionTag(txnID, tagID)` | `m.q.AddTransactionTag(ctx.Background(), ...)` |
| `s.RemoveTransactionTag(txnID, tagID)` | `m.q.RemoveTransactionTag(ctx.Background(), ...)` |
| `s.ListTransactionTags(txnID)` | `m.q.ListTransactionTags(ctx.Background(), txnID)` |
| `s.UpdateAccountBalance(id, bal)` | `m.q.UpdateAccountBalance(ctx.Background(), ...)` |
| `s.GetBaseCurrency()` | `shared.GetBaseCurrency()` |
| `s.Convert(amount, cur)` | `shared.Convert(amount, cur)` |
| `s.ctx()` | `context.Background()` |
| `s.recalculateBalance(id)` | `recalcBalance(m.q, id)` (method on Manager) |
| `parseDate(input)` | `shared.ParseDate(input)` |
| `parseMonth(input)` | `shared.ParseMonth(input)` |

### Error Types

`NotFoundError`, `ValidationError`, and sentinel errors move to `shared/shared.go`. All callers (CLI `helpers.go`, test files) update their references from `service.NotFoundError` / `service.ErrNotFound` to `shared.NotFoundError` / `shared.ErrNotFound`. No re-exports from `service.go` — that would just add indirection.

### CLI Changes (`internal/cli/helpers.go`)

Zero changes. `withService()` calls `getService()` which returns `*service.Service`. Embedded managers promote all methods — CLI command handlers see no difference.

### Test Changes

Each sub-package gets its own `_test.go` files alongside source files. Core test patterns:

- **Mock injection**: Pass `gen.Querier` mock to `NewManager(mockQuerier)` — same pattern as `NewWithQuerier`.
- **Rate config**: `shared.LoadRates` / `shared.SaveRates` vars allow test overrides (same pattern as current `SetTestRateConfig`).
- **Time mocking**: `todayFunc` stays in `forecast.go` (unchanged file).
- **`service_test.go` (3,476 lines)**: Tests for extracted methods move to their respective sub-package `_test.go` files. Tests for remaining `*Service` methods (account, category, tag, currency, forecast) stay in `service_test.go`. Tests that construct `Service` via `NewWithQuerier` continue passing since embedding promotes manager methods.

### Migration Order

1. Create `shared/` sub-package: move error types, resolvers, currency helpers, date helpers
2. Update `service.go`: add thin delegations to `shared/`, add manager embedding
3. Extract `transaction.go` → `transaction/` sub-package
4. Extract `planned_payment.go` → `plannedpayment/` sub-package
5. Extract `budget.go` → `budget/` sub-package
6. Extract `report.go` → `report/` sub-package
7. Delete old flat files
8. Run full test suite, fix any compilation errors
9. Run `make coverage-check` to verify 100% coverage maintained

### Naming

Sub-packages use lowercase, no underscores: `transaction`, `plannedpayment`, `budget`, `report`, `shared`.

File names within sub-packages describe the concern (e.g., `expense.go`, `rrule.go`, `breakdown.go`). No domain prefix needed — the package name scopes them.
