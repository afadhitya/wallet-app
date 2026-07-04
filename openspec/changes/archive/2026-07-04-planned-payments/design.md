## Context

The wallet app already has a SQLite-backed Go CLI with sqlc-generated queries, embedded migrations, service-layer validation, and Cobra commands for transactions, categories, tags, accounts, budgets, and balance recalculation. The initial data model already contains a `planned_payments` table concept, but there is no user-facing planned-payment workflow, no generated query surface for bills, and no transaction linkage behavior when a bill is paid.

Planned payments need to integrate with existing account/category validation and transaction persistence instead of introducing a separate payment ledger. Fulfillment remains manual: the app records a transaction only when the user runs `wallet bill pay <id>`.

## Goals / Non-Goals

**Goals:**

- Provide `wallet bill` commands for creating, listing, filtering due, paying, skipping, pausing, resuming, editing, and deleting planned payments.
- Store planned payments in SQLite and expose sqlc queries for all planned-payment operations.
- Link transactions created from bill payments back to their planned payment.
- Keep balance effects identical to normal expense transaction creation.
- Support one-time and recurring schedules with deterministic next-due calculations.
- Provide text-table and JSON output where the command contract requires it.
- Cover service behavior, recurrence edge cases, CLI output, and persistence with tests that satisfy the existing coverage policy.

**Non-Goals:**

- Automatic background charging or scheduled transaction creation.
- Notification delivery for upcoming bills.
- Bill tags or per-bill tag assignment.
- Rich RRULE editing UI beyond accepting and validating a custom recurrence rule string.
- Income-focused planned payments beyond preserving the existing model's payment type support; the CLI scope is bills/future expenses.

## Decisions

1. Use a dedicated `PlannedPaymentService` that composes existing query and transaction behavior.

   The service will live beside existing services and handle validation, due filters, state transitions, recurrence calculation, and pay/skip workflows. Paying a bill will create a normal expense transaction through the same persistence path used by transaction commands, then update the planned payment inside a single database operation where feasible. This keeps balance semantics consistent and avoids duplicating transaction rules in the CLI layer.

   Alternative considered: implement bill behavior directly in Cobra handlers. That would be faster initially but would duplicate validation and make service-level tests weaker.

2. Treat one-time bills as planned payments with recurrence `none`.

   A bill with only `--from` is stored with `recurrence = 'none'`, `start_date` and `next_due_date` set to the provided date. Paying it archives the planned payment by setting it inactive. Skipping a one-time bill is rejected because there is no next occurrence to advance to.

   Alternative considered: store one-time future expenses as pending transactions. That would blur the meaning of actual transactions and would require transaction lists to hide unpaid future records.

3. Store planned-payment fulfillment linkage on transactions.

   Transactions created by `wallet bill pay` will be marked with `is_planned = 1` and `planned_payment_id = <id>`. The fields make audit and future reporting possible without changing normal transaction behavior.

   Alternative considered: infer linkage from description or notes. That is fragile and not queryable enough for later forecasting or reconciliation work.

4. Compute simple recurrence schedules in application code and keep custom RRULE support isolated.

   Daily, weekly, monthly, and yearly recurrence can be computed with the standard `time` package. Monthly schedules use the existing due day as the intended day and clamp only when the target month is shorter. Custom RRULE parsing should be isolated behind a small function so a dependency can be added or replaced without affecting service callers.

   Alternative considered: store only RRULE values for every recurrence. That would make common cases harder to inspect and would introduce parsing complexity for simple schedules.

5. Soft-archive paid one-time bills and allow delete to remove planned-payment rows.

   One-time pay uses `is_active = 0` so historical linkage remains valid. `wallet bill rm` deletes a planned payment only when requested explicitly by the user. Due/list defaults exclude inactive rows so archived one-time bills do not clutter active workflows.

   Alternative considered: always delete one-time bills after pay. That would break transaction foreign-key linkage and remove useful audit history.

## Risks / Trade-offs

- Custom RRULE support could add dependency or parsing complexity -> isolate it behind a recurrence helper and add focused validation tests.
- Monthly date clamping can produce surprising dates if the implementation tries to restore the original day after a shorter month -> advance from the current due date and clamp only the immediate target month, so Jan 31 advances to Feb 28 and then Feb 28 advances to Mar 28.
- Paying a bill can partially write data if transaction creation and planned-payment advancement are not atomic -> perform create transaction, balance update, and planned-payment update in a transaction or use existing atomic service behavior if available.
- Existing schema may already include partial planned-payment columns -> inspect migrations before adding fields and only migrate missing linkage/state columns.
- Delete behavior can conflict with paid transaction linkage -> prefer archiving for paid one-time fulfillment and enforce referential behavior so deleting a planned payment does not corrupt existing transactions.

## Migration Plan

1. Inspect the current migration schema for `planned_payments` and `transactions` linkage columns.
2. Add a new migration only for missing planned-payment or transaction-linkage fields and indexes.
3. Add `internal/query/planned_payments.sql` and extend transaction insert/query definitions for planned linkage fields if needed.
4. Regenerate sqlc output and update services/CLI to use generated types.
5. Add tests before or alongside behavior implementation, including migration tests for the new schema shape.
6. Rollback by removing the new CLI/service/query code and reverting the migration before release; after release, preserve columns and disable commands instead of dropping linked data.

## Open Questions

- Whether duplicate payment prevention for the same bill period is required in this change or deferred; the brainstorming note mentions a possible future `--force` correction flow, but the command contract does not require it yet.
- Whether custom RRULE support should accept any RFC 5545 rule supported by the chosen parser or a documented subset.
