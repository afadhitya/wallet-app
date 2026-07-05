## Why

`wallet account list` sums raw balances across accounts regardless of currency, producing a meaningless total when accounts use different currencies. Additionally, the spec does not explicitly state that negative balances are allowed, despite the code already supporting them for credit cards, loans, and debt tracking.

## What Changes

- Convert each account's balance to the configured base currency using exchange rates before summing in `wallet account list`
- Display the total with a base-currency label (e.g., `Total (IDR): Rp 31.050.000`)
- Emit a warning when a rate is missing for an account's currency and exclude it from the total
- Update the `account-management-cli` spec to require currency-converted totals and explicitly allow negative balances

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `account-management-cli`: Account list total now converts all balances to the base currency before summing, and negative balances are explicitly allowed
- `multi-currency`: Extended mixed-currency display requirements to include account listing totals with base-currency conversion

## Impact

- `internal/cli/account.go` — `runAccountList()` needs currency conversion logic using `config.LoadRates()` or service `Convert()`
- `internal/cli/format.go` — may need a non-Rp currency prefix for the total display
- `openspec/specs/account-management-cli/spec.md` — delta spec for new total behavior and negative balance allowance
- `openspec/specs/multi-currency/spec.md` — delta spec for account listing display
- No database changes required
