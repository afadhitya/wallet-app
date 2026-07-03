## Context

The wallet app already has SQLite tables for budgets, budget category targets, and budget tag targets. It also has transaction entry/listing, category and tag resolution, sqlc-backed query packages, a service layer, and Cobra CLI commands. `wallet budget` currently exists only as a stub command.

Budgets are modeled as per-period snapshots. A recurring monthly budget is represented by separate budget rows for each period, linked to category and tag targets. Spending progress is calculated from existing expense transactions within the budget period.

## Goals / Non-Goals

**Goals:**
- Implement `wallet budget set`, `wallet budget list`, `wallet budget check`, `wallet budget edit <id>`, and `wallet budget rm <id>`.
- Support budgets that target categories, tags, or both.
- Calculate spent, remaining, percent used, and status from non-archived expense transactions.
- Auto-create a current recurring budget period on first check when a prior period exists and no current period exists.
- Provide text and JSON output paths consistent with existing CLI commands.
- Preserve the repository's quality expectation that generated sqlc code is regenerated and the Go coverage gate remains satisfied.

**Non-Goals:**
- Background notifications, schedulers, or daemon processes.
- Budget overlap prevention or deduplicating transactions that match multiple budgets.
- Multi-currency conversion for budget spending.
- Schema redesign of existing budget tables.
- Planned payment, reporting, or forecasting behavior.

## Decisions

1. Use the existing `budgets`, `budget_categories`, and `budget_tags` tables.
   - Rationale: the approved data model already defines budget snapshots and target links.
   - Alternative considered: add a separate budget definition table plus period table. That would be cleaner long-term, but it is larger than needed for the current model and would require a migration not justified by the MVP.

2. Treat budget `type` as the period kind for CLI inputs.
   - Rationale: the existing `budgets.type` column is available and the brainstorming input defines `monthly`, `weekly`, `yearly`, and `one_time` period choices. The service can store those values directly and use them for auto-calculation.
   - Alternative considered: store only `recurring` or `one_time`. That conflicts with the need to distinguish monthly, weekly, and yearly period boundaries.

3. Implement `budget set` as an upsert by name and current or explicit period.
   - Rationale: users expect setting the same named budget to update it rather than creating duplicates. Matching by name and period preserves historical period snapshots while making current-period edits straightforward.
   - Alternative considered: update by name across all periods. That would rewrite history and make recurring snapshots less meaningful.

4. Calculate category and tag spending separately and add the sums.
   - Rationale: the accepted behavior allows category and tag overlap and permits double-counting. Separate sqlc aggregate queries keep the rule explicit and simple.
   - Alternative considered: one distinct transaction query to avoid double-counting. That would violate the chosen overlap behavior and make mixed target behavior less transparent.

5. Soft-delete budgets by marking `is_active = 0` where practical.
   - Rationale: listing defaults to active budgets and the existing schema has an active flag. Preserving past budget rows is safer for historical review than physical deletion.
   - Alternative considered: physical deletion. It is simpler but discards historical budget definitions and target links.

6. Auto-generation copies the most recent prior period's amount and targets.
   - Rationale: this matches the recurring budget workflow without requiring a separate template entity.
   - Alternative considered: require users to run `budget set` every period. That adds friction and misses the recurring-budget objective.

## Risks / Trade-offs

- Budget `type` values may differ from earlier schema examples that used `recurring` → Store documented period values consistently and cover them in service tests.
- Auto-generation based on the current date can make tests flaky → Keep period calculation in small helpers and test with deterministic dates where possible.
- Mixed category and tag targets can double-count a transaction → Document this in specs and keep tests aligned with the accepted behavior.
- Soft deletion may require query updates for list/check behavior → Make active filtering explicit in budget queries and CLI flags.
- JSON/text output can diverge → Test both output modes for representative commands.

## Migration Plan

No schema migration is expected because budget storage tables already exist. Implementation should add budget sqlc query files, regenerate `internal/gen`, add service and CLI code, and extend tests. Rollback is code-only: remove the budget command implementation and queries while leaving existing schema untouched.

## Open Questions

None.
