## Why

Budget `calculateSpending` sums category-based and tag-based expenses independently and adds them together. A single transaction matching both targets (e.g., a transaction in a budget's linked category AND bearing a budget's tracked tag) is counted twice, producing inflated spent amounts and unreliable budget check/list/notify thresholds.

## What Changes

- Replace the two independent `SumCategoryExpenses` + `SumTagExpenses` queries in `calculateSpending` with a single deduplicated query using `OR`-based UNION logic.
- Remove or regenerate the now-unused `SumCategoryExpenses` and `SumTagExpenses` generated queries; add a replacement `SumBudgetExpenses` query that deduplicates overlapping transactions.

## Capabilities

### Modified Capabilities
- `budget-engine`: Replace the double-counting spending calculation with deduplicated spending. The existing "Mixed target overlap is double-counted" spec scenario changes from "system does not deduplicate" to "system deduplicates overlapping transactions so each transaction is counted at most once."

## Impact

- `internal/query/budgets.sql` — replace `SumCategoryExpenses` and `SumTagExpenses` with a single `SumBudgetExpenses` query.
- `internal/gen/` — regenerate sqlc output (removes two old query methods, adds one new one).
- `internal/service/budget.go` — replace `calculateSpending` implementation to call the new single query.
- `internal/service/budget_test.go` — update tests to verify deduplication behavior.
