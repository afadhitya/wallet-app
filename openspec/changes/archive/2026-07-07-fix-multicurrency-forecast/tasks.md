## 1. Core Implementation — Currency Conversion in ForecastBalance

- [x] 1.1 Convert account balances to base currency in the start-balance aggregation loop (`internal/service/forecast.go:180-182`): replace `startBalance += acc.Balance` with `s.Convert(acc.Balance, acc.Currency)`, loading rate config once before the loop
- [x] 1.2 Convert planned payment amounts in the per-month projection loop (`internal/service/forecast.go:189-196`): use `s.Convert(mo.occ.Amount, mo.occ.Currency)` instead of raw `mo.occ.Amount`, skip payments whose currency lacks a configured rate and append a warning
- [x] 1.3 Convert planned payment amounts in the category breakdown loop (`internal/service/forecast.go:225-230`): use converted amounts for `categoryExpenses` totals to ensure consistency with monthly projections
- [x] 1.4 Add a `warnings` entry when a planned payment is skipped due to a missing exchange rate, identifying the payment name and missing currency

## 2. Tests — Multi-Currency Forecast Verification

- [x] 2.1 Add `TestForecastBalance_MultiCurrencyStartBalance` — verify that start balance correctly converts USD and JPY accounts to IDR using configured rates (e.g., USD rate 15800, JPY rate 105), confirming the aggregate equals converted sum, not raw sum
- [x] 2.2 Add `TestForecastBalance_MultiCurrencyMonthlyProjection` — verify that planned payments on foreign-currency accounts contribute converted amounts to monthly income/expenses
- [x] 2.3 Add `TestForecastBalance_MissingRateSkipsPayment` — verify that a planned payment whose currency lacks a configured rate is excluded from projections and generates a warning
- [x] 2.4 Verify existing single-currency forecast tests (`TestForecastBalance_DefaultOneMonth`, `TestForecastBalance_MultiMonth`, `TestForecastBalance_WithAccountFilter`, etc.) still pass unchanged

## 3. Validation

- [x] 3.1 Run `make fmt` to format code
- [x] 3.2 Run `make lint` (golangci-lint) to verify no linter issues
- [x] 3.3 Run `make test` to confirm all tests pass
- [x] 3.4 Run `make coverage-check` to confirm coverage threshold is met
