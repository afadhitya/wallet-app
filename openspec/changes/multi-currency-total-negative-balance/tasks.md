## 1. Implement Currency-Converted Total

- [ ] 1.1 Import `config` package and `fmt` in `internal/cli/account.go`
- [ ] 1.2 Load `config.RateConfig` in `runAccountList()` via `config.LoadRates()`
- [ ] 1.3 Convert each non-base-currency account balance using `rateCfg.Rates[account.Currency]` before adding to `totalBalance`
- [ ] 1.4 Track missing rates and print a warning when a currency has no configured rate
- [ ] 1.5 Update total row label to include base currency (e.g., `Total (IDR):` instead of `Total`)

## 2. Update Tests

- [ ] 2.1 Add test for mixed-currency account list total (e.g., IDR + USD accounts with `SetTestRateConfig`)
- [ ] 2.2 Add test for missing rate warning when an account's currency has no configured rate
- [ ] 2.3 Add test for account list with negative balances included in total
- [ ] 2.4 Verify existing account list tests still pass

## 3. Verification

- [ ] 3.1 Run `make test` and ensure all tests pass
- [ ] 3.2 Run `make coverage-check` and ensure 100% coverage
- [ ] 3.3 Run `make lint` and fix any issues
