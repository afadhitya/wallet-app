## Context

The wallet app already stores accounts with integer balances and planned payments with type, amount, recurrence, active state, paused state, account, category, and next due date. Planned-payment commands can list due payments and fulfill bills, but the current `forecast` command is only a placeholder and does not project future balances.

Forecasting should use planned payments as the only projection source. This keeps results predictable, avoids trying to infer spending from historical transactions, and lets users improve forecast accuracy by maintaining recurring income and bills.

## Goals / Non-Goals

**Goals:**
- Add `wallet forecast` for monthly balance projection over a configurable horizon.
- Add `wallet forecast bills` for upcoming bill impact over a configurable horizon.
- Support `--months`/`-n`, `--account`/`-a` for balance forecasts, and global `--json` output.
- Reuse active, unpaused planned payments and current account balances.
- Warn, but do not fail, when a projected ending balance becomes negative.

**Non-Goals:**
- Forecast unplanned or historical average spending.
- Add tag-level forecast breakdowns.
- Add new database tables or persisted forecast snapshots.
- Support currency conversion across accounts; forecasts use each planned payment and account currency as currently stored.

## Decisions

- Forecast from planned payments only.
  - Rationale: planned payments are explicit user intent and already contain due dates, recurrence, account, category, type, and amount.
  - Alternative considered: infer future spend from transaction history. This would be less predictable and would require additional heuristics outside this change.

- Expand recurring planned payments in service code across the requested horizon.
  - Rationale: SQL can filter seed payments by due date, but recurrence expansion is easier to keep consistent with existing planned-payment date logic in Go.
  - Alternative considered: aggregate directly in SQL by `next_due_date` only. That would miss future recurring occurrences beyond the stored next due date.

- Implement forecasting as methods on the existing `service.Service` type.
  - Rationale: the current app uses one service facade with generated sqlc queries, account/category resolution helpers, and CLI integration through `withService`.
  - Alternative considered: introduce a separate `ForecastService`. That adds indirection without clear reuse benefits in the current codebase.

- Use monthly buckets for balance forecasts and dated rows for bill forecasts.
  - Rationale: `wallet forecast --months N` needs a concise multi-month projection, while `wallet forecast bills` needs ordered due dates and running totals.
  - Alternative considered: one detailed daily ledger for all forecasts. That would be noisier and beyond the requested command shape.

- Keep missing planned payments as a successful empty result.
  - Rationale: no forecast data is an actionable state, not an application failure. Text output should guide the user to add bills; JSON output should expose empty collections and zero totals.

## Risks / Trade-offs

- Recurrence expansion can diverge from payment fulfillment logic if implemented separately -> reuse or extract existing recurrence helpers where practical and cover month-end clamping in tests.
- Multi-account forecasts with mixed currencies can be misleading if totals are combined -> present per-account/month rows or make clear that values are in stored account currencies; do not introduce conversion in this change.
- Forecasts become inaccurate when users omit planned payments -> text output should state that forecasts are based on planned payments only.
- Negative balances are warnings, not failures -> include warnings in text and JSON so automation can inspect them without non-zero exits.
