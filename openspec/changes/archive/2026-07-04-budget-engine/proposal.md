## Why

Wallet users can record and classify spending, but they cannot define limits or see progress against those limits. Adding budget management turns the existing category, tag, and transaction data into actionable spending controls for recurring and one-time goals.

## What Changes

- Add budget creation and update support through `wallet budget set`, including category and tag targets, period selection, notification threshold, and JSON output.
- Add budget listing through `wallet budget list`, showing active budgets by default with spent and remaining amounts.
- Add budget checking through `wallet budget check`, calculating spending for targeted expense transactions and reporting `ok`, `warning`, or `over` status.
- Add budget editing through `wallet budget edit <id>` for amount, name, notification threshold, categories, and tags.
- Add budget deletion through `wallet budget rm <id>`.
- Auto-create recurring budget periods on first check when the current period does not yet exist, copying the previous period's limit and targets.
- Add sqlc queries, service-layer behavior, CLI rendering, validation, and tests for budget management.

## Capabilities

### New Capabilities
- `budget-engine`: Budget management commands, service behavior, spending calculation, recurring period generation, validation, and test expectations.

### Modified Capabilities

## Impact

- Affected CLI: budget command group under `wallet budget` with `set`, `list`, `check`, `edit`, and `rm` subcommands.
- Affected services: new budget service logic using existing transaction, category, tag, and budget tables.
- Affected database access: new sqlc queries for budgets, budget targets, spending totals, and period lookup.
- Affected tests: service unit tests, CLI integration tests, JSON/text output tests, validation/error tests, and coverage gate compliance.
