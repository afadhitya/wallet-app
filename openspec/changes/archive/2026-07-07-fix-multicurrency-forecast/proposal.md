## Why

When users hold accounts in multiple currencies (e.g., IDR, USD, JPY), `wallet forecast` sums raw account balances without converting foreign-currency amounts to the base currency. This produces a meaningless aggregate that severely underestimates net worth for users with non-IDR accounts, showing premature negative balances and unreliable projections.

## What Changes

- Convert each account's balance to the base currency (IDR) using configured exchange rates before summing the start balance when no account filter is provided
- Convert planned payment amounts to the base currency before aggregating projected income/expenses per month
- Skip planned payments whose account currency lacks a configured exchange rate, logging a warning

## Capabilities

### New Capabilities
<!-- None — this is a bug fix within existing capabilities -->

### Modified Capabilities
- `multi-currency`: The mixed-currency aggregation requirement already covers account listing and reports. This change extends that requirement to the forecast start-balance and monthly projection aggregation.
- `forecasting`: The start balance calculation scenario must now convert non-base-currency accounts before summing. The forecast breakdown must aggregate planned payment amounts in base currency.

## Impact

- **Code**: `internal/service/forecast.go` — `ForecastBalance` function (start balance loop and monthly projection loop)
- **Tests**: `internal/service/forecast_test.go` — add multi-currency test cases with USD/JPY accounts
- **No breaking changes**: The fix corrects incorrect behavior; single-currency users are unaffected
