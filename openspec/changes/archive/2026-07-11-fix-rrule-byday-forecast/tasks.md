## 1. Implement BYDAY parsing in calcNextDueFromRRULE

- [x] 1.1 Add `byday` weekday lookup map (MO→Monday, TU→Tuesday, etc.) in `planned_payment.go`
- [x] 1.2 Parse `BYDAY` values from the RRULE string parts in `calcNextDueFromRRULE`
- [x] 1.3 Implement weekday advance logic: starting from `currentDue + 1 day`, loop up to 7 days to find the first matching weekday
- [x] 1.4 Fall back to plain `+7 days` when no `BYDAY` is present (backward compatibility)
- [x] 1.5 Run `make test` to verify existing tests still pass

## 2. Add unit tests for BYDAY scenarios

- [x] 2.1 Add `TestCalcNextDue_RRULEWeeklyBYDAY_SingleDay` — single `BYDAY=WE` on a Tuesday advances to Wednesday
- [x] 2.2 Add `TestCalcNextDue_RRULEWeeklyBYDAY_MultiDay` — `BYDAY=TU,WE` on Saturday advances to Tuesday
- [x] 2.3 Add `TestCalcNextDue_RRULEWeeklyBYDAY_SameDay` — `BYDAY=TH` on Thursday advances to next Thursday
- [x] 2.4 Add `TestCalcNextDue_RRULEWeeklyBYDAY_LastDayOfWeek` — `BYDAY=SU` on Sunday advances to next Sunday
- [x] 2.5 Add `TestCalcNextDue_RRULEWeekly_NoBYDAY` — verify plain `FREQ=WEEKLY` still advances by 7 days
- [x] 2.6 Run `make coverage-check` to confirm 100% coverage maintained

## 3. Verify end-to-end behavior

- [x] 3.1 Run `make build` to ensure clean compilation
- [x] 3.2 Run `make test` to confirm all tests pass
