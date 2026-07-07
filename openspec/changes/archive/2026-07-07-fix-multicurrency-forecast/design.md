## Context

The `wallet forecast` command aggregates account balances and planned payment amounts when projecting future balances. Currently, the `ForecastBalance` service method sums `int64` account balances without converting non-base-currency accounts to the base currency (IDR). The service already has a fully functional `Convert(amount, fromCurrency)` method in `internal/service/currency.go` that handles this conversion — it simply isn't called in the forecast code path.

The same aggregation was already done correctly in `wallet account list` (via SQL with `base_amount` columns or in-code conversion in `account_service.go`), so the pattern is established.

## Goals / Non-Goals

**Goals:**
- Convert each account's balance to the base currency before summing the start balance when no `--account` filter is provided
- Convert planned payment amounts before aggregating per-month projected income/expenses
- Skip planned payments whose account currency has no configured rate, with a warning
- Maintain backward compatibility: single-currency users (all IDR accounts) see identical behavior

**Non-Goals:**
- Changing how individual account-scoped forecasts work (when `--account` is specified, the start balance is a single account's balance — no aggregation needed)
- Changing the display format of forecast output
- Adding the base-currency label to forecast output (already labeled as the base currency)
- Modifying the stored exchange rate data model or TOML format

## Decisions

### Decision 1: Convert in the `ForecastBalance` service method, not in SQL

**Rationale:** The start balance aggregation already happens in Go (a loop over `accounts`), not via a sqlc query. The `Convert` method is a pure Go function on the `Service` struct. Adding a SQL query that joins `accounts` with `exchange_rates` would be more invasive and less testable. Keeping the conversion in Go leverages existing patterns (`Convert` is already used in `transaction.go` for `resolveBaseFields`).

**Alternatives considered:**
- SQL-based aggregation: Would require a new sqlc query, redefining the projection logic in SQL. Rejected because the conversion is simple arithmetic and the existing forecast logic is already Go-based.
- Pre-computing base balances on the `accounts` table: Would change the data model and require migration, which is overkill for a display/aggregation concern.

### Decision 2: Skip planned payments with missing rates, log a warning

**Rationale:** Failing hard would break the forecast for users who have a planned payment in a currency they haven't yet configured. The forecast is a projection tool, not a transactional operation. Skipping with a warning is the same pattern used by `wallet account list` for unconfigured currencies.

### Decision 3: Do not modify the `ForecastBalance` signature

**Rationale:** The return type (`[]ForecastRow`) already contains `int64` amounts. These will now represent base-currency amounts when aggregating multiple accounts. Single-account forecasts (via `--account`) already returned that account's raw balance, which is correct because the display layer handles formatting. No signature change needed.

## Risks / Trade-offs

- **Risk:** Converting large JPY balances (e.g., ¥100,000) using a rate of 105 produces ~10,500,000 IDR, which could overflow `int64` in extreme edge cases. **Mitigation:** This is the same risk as existing transaction conversion (`resolveBaseFields`). The `int64` range (±9.2e18) is well beyond realistic personal finance amounts.

- **Risk:** If a user has a mismatch between their account currency and the rate config (typo in currency code), the conversion silently uses raw balance (base currency identity). **Mitigation:** This is existing behavior. The forecast faithfully reflects what the conversion layer returns.
