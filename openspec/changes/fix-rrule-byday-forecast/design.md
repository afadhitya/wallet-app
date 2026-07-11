## Context

The `calcNextDueFromRRULE` function in `internal/service/planned_payment.go:604` parses custom RRULE strings to compute the next due date for recurring planned payments. Currently it parses `BYMONTHDAY` but ignores `BYDAY`. For `FREQ=WEEKLY` with `BYDAY=TU,WE`, the function falls through to the plain WEEKLY case and adds 7 days unconditionally.

This function is the single source of truth for advancing dates — it is called from:
- `calcNextDue` (which routes custom recurrence here)
- `expandOccurrences` in the forecast layer (which iteratively calls `calcNextDue`)

Both `wallet bill pay`/`skip` (single advance) and `wallet forecast` (iterative expansion) depend on this function.

## Goals / Non-Goals

**Goals:**
- Parse `BYDAY` values from custom RRULE strings when `FREQ=WEEKLY`
- Compute the next due date as the earliest upcoming date matching one of the specified weekdays
- Handle multiple `BYDAY` values (comma-separated, e.g., `TU,WE`)
- Preserve existing behavior when `BYDAY` is absent or for non-WEEKLY frequencies

**Non-Goals:**
- Adding a full RFC 5545 RRULE parser (e.g., BYMONTH, BYSETPOS, INTERVAL, COUNT, UNTIL)
- Changing the data model or database schema
- Adding support for weekly recurrence with `BYDAY` as a top-level recurrence option (stays under custom)

## Decisions

### Decision 1: Extend the existing RRULE parser rather than add a library dependency

**Rationale:** The current `calcNextDueFromRRULE` already does manual string parsing for `FREQ` and `BYMONTHDAY`. Adding a full RFC 5545 library (e.g., `github.com/teambition/rrule-go`) would be overkill for parsing one extra parameter. The function is ~45 lines and adding `BYDAY` parsing adds ~20 more.

**Alternatives considered:**
- Third-party RRULE library — adds a dependency, changes behavior for existing rules, harder to reason about edge cases. Rejected as disproportionate.
- Refactor entire recurrence system to use a proper library — valuable long-term but out of scope for this bug fix.

### Decision 2: Parse `BYDAY` values into `time.Weekday` and find the next matching day

**Rationale:** `BYDAY` values are two-character abbreviations (MO, TU, WE, TH, FR, SA, SU) that map directly to `time.Weekday`. Parsing is straightforward with a lookup map. The algorithm: given the current due date, find the next date matching any of the specified weekdays.

For `BYDAY=TU,WE` with a current due date of Saturday July 4, the next occurrence is Tuesday July 7.

### Decision 3: Handle the `BYDAY` lookup with a simple weekday advance loop

**Rationale:** After parsing `BYDAY` values into a `map[time.Weekday]bool`, starting from `currentDue + 1 day`, loop up to 7 days forward. The first date whose weekday is in the map is the next occurrence. This is deterministic, cheap (max 7 iterations), and easy to test.

**Edge case:** If no `BYDAY` value is provided (plain `FREQ=WEEKLY`), fall through to the existing `+7 days` behavior — zero breaking change.

## Risks / Trade-offs

- [BYDAY with no matching days parsed] The parser silently ignores unrecognized BYDAY tokens. If BYDAY is provided but parsing fails (e.g., garbled input), the function falls back to plain weekly +7. Mitigation: the `validateRRULE` function already validates the RRULE structure; BYDAY tokens use a well-defined 2-letter format.
- [Performance] The loop iterates at most 7 times — negligible. No performance risk.
- [Single-day BYDAY vs multi-day] `BYDAY=WE` should advance to next Wednesday even if current date is a Wednesday — consistent with how recurrence works (the due date is the past occurrence, next is the future one). The loop starts from `currentDue + 1 day`, so it correctly finds the *next* Wednesday.
