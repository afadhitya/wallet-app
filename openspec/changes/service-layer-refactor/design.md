## Context

The `internal/service/` directory currently contains flat `.go` files — one per domain. Four files exceed 500 lines, with `transaction.go` reaching 753 lines. This makes navigation, reasoning, and test organization difficult. The project uses a clean architecture with presenters (CLI), business logic (service), and data access (sqlc-generated `gen.Querier`). The refactor must preserve the public API surface (`*Service` methods) and maintain 100% test coverage.

## Goals / Non-Goals

**Goals:**
- Split bloated domain files into focused sub-packages with files per concern
- Introduce a `shared/` sub-package for cross-cutting utilities to avoid circular imports
- Preserve the `*Service` API surface via struct embedding (zero caller changes)
- Maintain 100% test coverage
- Keep the `gen.Querier` interface as the sole data access mechanism

**Non-Goals:**
- Changing business logic, validation rules, or output formats
- Modifying CLI command signatures or JSON output structure
- Introducing new external dependencies
- Changing the database schema or sqlc-generated code
- Refactoring non-bloated service files (account.go, category.go, tag.go, currency.go, forecast.go)

## Decisions

### Manager Struct Pattern
Each sub-package defines a `Manager` struct holding a `gen.Querier` and exposes a `NewManager(q gen.Querier) *Manager` constructor. All domain methods become methods on `*Manager`. This is a natural Go pattern for grouping related operations and aligns with the existing `gen.Querier`-based testability strategy.

### Service Composition via Embedding
Rather than writing delegation methods for every extracted method, the `Service` struct embeds `*Manager` pointers:

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

Embedding promotes all `Manager` methods to `*Service` automatically. **Alternative considered**: explicit delegation methods for each method. Rejected because it adds ~40+ lines of boilerplate per domain with no benefit — callers see the same methods either way.

### Cross-Call Migration Strategy
Extracted methods that previously called other `*Service` methods are updated to use their replacements:

| Old call | Replacement |
|---|---|
| `s.ResolveCategory(name)` | `shared.ResolveCategory(m.q, name)` |
| `s.ctx()` | `context.Background()` |
| `s.GetBaseCurrency()` | `shared.GetBaseCurrency()` |
| `s.recalculateBalance(id)` | `m.recalcBalance(id)` (same method, now on Manager) |

The `Service` struct retains thin delegation methods (`s.ResolveCategory(...)` → `shared.ResolveCategory(s.q, ...)`) for the non-extracted service files (forecast, account, category, tag, currency) that continue calling them.

### Error Type Location
`NotFoundError`, `ValidationError`, and sentinel errors move to `shared/shared.go`. Callers update imports from `service.NotFoundError` to `shared.NotFoundError`. No re-exports — that would just add indirection and defeat the purpose of the shared package.

### File Split Strategy
Each extracted domain splits into files by concern rather than by code size:

- **transaction/**: `manager.go`, `types.go`, `expense.go`, `income.go`, `transfer.go`, `adjustment.go`, `edit.go`, `list.go`
- **plannedpayment/**: `manager.go`, `types.go`, `create_edit.go`, `pay_skip.go`, `states.go`, `rrule.go`
- **budget/**: `manager.go`, `types.go`, `create.go`, `check.go`, `spending.go`, `edit.go`
- **report/**: `manager.go`, `types.go`, `breakdown.go`, `export.go`

Types (params, results) live in `types.go` per sub-package. Manager struct and constructor in `manager.go`.

### Migration Order
The shared package must be created first since all sub-packages depend on it. Individual domain extractions can happen in any order but are applied sequentially to keep diffs incremental. Each extraction includes moving corresponding tests.

## Risks / Trade-offs

- **Import path changes for CLI/tests**: `service.NotFoundError` → `shared.NotFoundError`. This is a mechanical search-and-replace across `internal/cli/` and test files. Risk: missed references cause compile errors → Mitigation: `go build` catches all; run full test suite after each extraction step.

- **Test file split complexity**: `service_test.go` is 3,476 lines. Extracting tests per sub-package requires careful identification of which tests belong where. Risk: tests left behind or duplicated → Mitigation: use `grep` to identify test functions referencing extracted methods; verify with `go test ./...` after each move.

- **Circular import prevention**: `shared/` must never import `service/` or any sub-package. Risk: accidental import during refactoring → Mitigation: this is a structural rule enforced by `go build`; the compiler rejects circular imports.

- **Coverage tracking**: Previously all service code was in one package. After split, coverage is tracked per package. Risk: `coverignore.txt` may need updates for new file paths → Mitigation: verify `make coverage-check` after refactor; update `coverignore.txt` if needed.
