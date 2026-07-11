## Why

The `calcNextDueFromRRULE` function in the service layer does not parse `BYDAY` from custom RRULE strings. When a weekly recurring bill specifies `BYDAY=TU,WE`, the forecast command ignores the day-of-week constraint and instead generates dates every 7 days from the start date, showing incorrectly scheduled bill occurrences.

## What Changes

- Parse `BYDAY` values from custom RRULE strings in `calcNextDueFromRRULE` when the frequency is `WEEKLY`
- Compute the next occurrence based on the specified days of the week rather than blindly adding 7 days
- Handle multiple `BYDAY` values (e.g., `BYDAY=TU,WE`) by advancing to the next matching weekday
- Maintain backward compatibility: plain `FREQ=WEEKLY` (no `BYDAY`) continues to advance by 7 days

## Capabilities

### New Capabilities
- None. This is a bug fix to existing recurrence calculation behavior.

### Modified Capabilities
- `planned-payments`: The recurrence calculation requirement is updated to specify that `BYDAY` in WEEKLY custom RRULE strings must influence the next due date computation, not just a flat 7-day increment.

## Impact

- `internal/service/planned_payment.go` — `calcNextDueFromRRULE` function (line 604)
- `internal/service/planned_payment_test.go` — new test cases for BYDAY variants
- `internal/service/forecast.go` — indirectly affected; `expandOccurrences` calls `calcNextDue` which routes to `calcNextDueFromRRULE` for custom recurrence
