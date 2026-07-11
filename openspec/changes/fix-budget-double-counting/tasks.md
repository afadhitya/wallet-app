## 1. SQL Query

- [ ] 1.1 Replace `SumCategoryExpenses` and `SumTagExpenses` in `internal/query/budgets.sql` with a single `SumBudgetExpenses` query using LEFT JOIN + OR deduplication
- [ ] 1.2 Run `sqlc generate` (or project's regenerate command) to update `internal/gen/`

## 2. Service Layer

- [ ] 2.1 Update `calculateSpending` in `internal/service/budget.go` to call `SumBudgetExpenses` instead of the two old queries
- [ ] 2.2 Remove unused `toInt64` helper if no longer referenced

## 3. Tests

- [ ] 3.1 Update service tests in `internal/service/budget_test.go` to cover the deduplication scenario (a transaction matching both category and tag targets is counted once)
- [ ] 3.2 Remove or update any test assertions referencing `SumCategoryExpenses` or `SumTagExpenses`

## 4. Verification

- [ ] 4.1 Run `go build ./...` to confirm compilation
- [ ] 4.2 Run `go test ./internal/service/...` and fix any failures
- [ ] 4.3 Run full test suite and coverage check per project conventions
