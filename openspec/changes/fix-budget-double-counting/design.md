## Context

`calculateSpending` in `internal/service/budget.go` calls two independent sqlc-generated queries (`SumCategoryExpenses` and `SumTagExpenses`) and adds their results. These queries scan independent JOIN paths — one through `budget_categories`, the other through `transaction_tags` → `budget_tags` — with no deduplication. A transaction satisfying both paths is counted twice.

The fix replaces both queries with a single `SumBudgetExpenses` query that uses `LEFT JOIN` with `OR` logic, so each matching expense transaction contributes at most once.

## Goals / Non-Goals

**Goals:**
- Each expense transaction matching a budget's targets is counted exactly once in that budget's spent amount.
- Minimal change: one new SQL query, regenerate sqlc, update `calculateSpending`.
- Existing budget behavior unchanged except for overlap deduplication.

**Non-Goals:**
- Changing the budget data model (`budget_categories`, `budget_tags` tables).
- Changing how budgets are set, edited, listed, or removed.
- Changing non-expense transaction handling (already excluded).

## Decisions

### Decision 1: Single query with `LEFT JOIN` + `OR` over UNION

**Chosen**: Single query pattern:
```sql
SELECT COALESCE(SUM(t.amount), 0)
FROM transactions t
LEFT JOIN budget_categories bc ON bc.budget_id = ? AND bc.category_id = t.category_id
LEFT JOIN transaction_tags tt ON tt.transaction_id = t.id
LEFT JOIN budget_tags bt ON bt.budget_id = ? AND bt.tag_id = tt.tag_id
WHERE t.type = 'expense' AND t.is_archived = 0
  AND t.date >= ? AND t.date <= ?
  AND (bc.budget_id IS NOT NULL OR bt.budget_id IS NOT NULL)
```

**Alternatives considered**:
- `UNION` of two subqueries with `SELECT DISTINCT t.id` — more verbose, same result, extra sqlc naming complexity.
- In-app dedup via Go code (fetch IDs from both queries, compute union, sum) — unnecessary app logic, possible N+1, harder to maintain.

**Rationale**: The `LEFT JOIN` + `OR` approach is a single query, naturally deduplicates because each row yields once from the single `SUM(t.amount)`, and follows the existing sqlc query pattern.

### Decision 2: Remove old queries rather than deprecate

**Chosen**: Remove `SumCategoryExpenses` and `SumTagExpenses` from `budgets.sql`, regenerate, and update the single caller.

**Rationale**: These queries are only called from `calculateSpending`. No other code references them. Deleting is cleaner than leaving dead code.

## Risks / Trade-offs

- **Risk**: `LEFT JOIN` on three tables may perform worse than two targeted `INNER JOIN` queries on very large datasets. → Mitigation: SQLite query planner handles this pattern well; budget queries run on small per-period subsets already filtered by date range. If performance becomes measurable, add composite indexes.
- **Risk**: Regenerated code removes `SumCategoryExpenses` and `SumTagExpenses` from the `Querier` interface — any external mocks or tools relying on the interface will break. → Mitigation: The `Querier` interface is generated and only used within this codebase. Service tests use the real in-memory DB or a mock generator. Update test file accordingly.
