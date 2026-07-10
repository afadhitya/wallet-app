## Context

The `transactions` table has a `planned_payment_id` foreign key column and an `is_planned` boolean column. The FK references `planned_payments(id)` with no `ON DELETE` action. When `bill rm` hard-deletes a planned payment that has been paid, SQLite rejects the operation because the linked transaction still references it.

The columns were added for "audit and future reporting" but are never read — no report, budget, statistics, or CLI query joins transactions back to planned payments via these columns.

## Goals / Non-Goals

**Goals:**
- Remove `planned_payment_id` and `is_planned` from the `transactions` table
- Allow `bill rm` to hard-delete paid planned payments without FK errors
- Keep transactions created by `bill pay` as independent records

**Non-Goals:**
- Preserve the planned payment link on existing transactions (nothing reads it)
- Enable rollback of the migration (consistent with existing forward-only migration strategy)
- Change how `bill pay` creates transactions (same amount, type, category, account behavior)

## Decisions

1. **DROP COLUMN via migration** — SQLite 3.35.0+ supports `ALTER TABLE DROP COLUMN`. The project uses `modernc.org/sqlite` (pure Go), which supports this. This is cleaner than a multi-step workaround (create new table, copy data, drop old, rename).

2. **Remove `CreatePlannedTransaction` query entirely** — Instead of repurposing it, delete it and use the existing `CreateTransaction`. This reduces generated interface surface area and keeps the query layer simple.

3. **No data preservation** — The `planned_payment_id` link on past transactions is discarded. Since nothing reads it, preserving it adds complexity with zero value.

4. **Plain transaction on pay** — `PayPlannedPayment` creates a standard expense/income/transfer via `CreateTransaction` with no special marking. The transaction is indistinguishable from one created manually via `wallet add`.

## Risks / Trade-offs

- **[No rollback]**: `DROP COLUMN` is irreversible in SQLite. Migration is forward-only, consistent with the existing strategy. Acceptable since the columns provide no value.
- **[No FK safety net]**: After removal, deleting a planned payment no longer cascades or blocks on linked transactions. This is intentional — transactions should survive bill deletion.
