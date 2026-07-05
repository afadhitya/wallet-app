## Context

`runAccountList()` in `internal/cli/account.go` sums `acc.Balance` across all accounts without currency conversion. A USD account with 1,000 USD and an IDR account with 15,000,000 IDR produces a total of 15,001,000 — a meaningless number.

The service layer already has `Convert()`, `GetBaseCurrency()`, and `loadRateConfig()` methods. However, these use `math.Round(float64(amount) * float64(rate))` which introduces unnecessary float conversion. For display totals, direct integer multiplication is simpler and avoids precision issues.

Negative balances already work at the schema and service level. No code changes needed — only spec documentation.

## Goals / Non-Goals

**Goals:**
- Fix `wallet account list` total to convert foreign-currency balances to the base currency before summing
- Display the total with a base-currency label (e.g., `Total (IDR):`)
- Warn when an account's currency has no configured rate and exclude it from the total
- Update specs to document currency-converted totals and explicit negative balance allowance

**Non-Goals:**
- Changing how individual account balances are displayed (still `formatAmount()` per row)
- Changing transaction-level currency conversion logic
- Adding real-time exchange rate fetching
- Changing the service layer's `Convert()` method

## Decisions

### D1: Rate Loading Strategy

**Choice:** Load `config.RateConfig` directly in `runAccountList()` via `config.LoadRates()`.

**Rationale:** The service layer's `Convert()` uses `math.Round(float64 * float64)` which is unnecessary for display totals. Direct integer multiplication preserves precision and avoids an unnecessary dependency on the service for a display concern. The CLI already imports `pkg/config` in other commands.

**Alternative considered:** Use `svc.GetBaseCurrency()` + `svc.Convert()`. Rejected because it introduces float rounding for a display total and couples CLI display logic to service methods designed for transaction processing.

### D2: JSON Output Behavior

**Choice:** JSON output includes the raw list of accounts unmodified. The converted total is only relevant for table display.

**Rationale:** JSON consumers can perform their own currency conversion. Adding a `total_converted` field would be a separate API concern. The existing JSON output contract — returning the account list — remains unchanged.

### D3: Warning for Missing Rates

**Choice:** Print a warning to stderr listing the missing currency codes, and exclude those accounts from the total.

**Rationale:** The total should be meaningfully accurate. Including unconverted foreign balances in the total is misleading. A warning gives the user a clear action: add the missing rate.

## Risks / Trade-offs

- **Rate recency**: If the user updates rates after viewing the list, the total changes. This is expected behavior — the total reflects current configured rates, not historical rates. No mitigation needed.
- **Large integer overflow**: `int64` multiplication of balance × rate could theoretically overflow for extreme values (e.g., balance in billions times a large rate). In practice, personal finance balances and rates stay well within int64 range. No special overflow handling needed.
- **Performance**: `config.LoadRates()` reads a small TOML file from disk on every `wallet account list` invocation. Acceptable for a CLI tool.
