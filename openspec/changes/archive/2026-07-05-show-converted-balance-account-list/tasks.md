## 1. Add Converted Column to Table Output

- [x] 1.1 Update table header format string in `runAccountList` (`internal/cli/account.go:145-146`) to include "Converted" column between "Balance" and "Status", with matching separator dashes
- [x] 1.2 Add per-row converted balance logic in the accounts loop (`internal/cli/account.go:150-165`): compute `convertedStr` as `-` for base-currency and missing-rate accounts, `formatAmount(balance * rate)` for non-base-currency accounts with a rate
- [x] 1.3 Update the bottom separator row format string (`internal/cli/account.go:168`) to include the extra column width for alignment with the new header

## 2. Update Tests

- [x] 2.1 Update `TestCLIAccountList` in `internal/cli/account_test.go` to assert the new "Converted" column appears in table output with `-` for IDR (base-currency) accounts
- [x] 2.2 Update `TestCLIAccountListMixedCurrencies` to assert non-base-currency accounts show converted amounts in the "Converted" column, base-currency accounts show `-`
- [x] 2.3 Update `TestCLIAccountListMissingRate` to assert accounts with missing rates show `-` in the "Converted" column
- [x] 2.4 Update `TestCLIAccountListNegativeBalances` to assert negative balances produce negative converted amounts
- [x] 2.5 Update `TestCLIAccountListAll` to include converted column assertions
- [x] 2.6 Update integration test `TestCLIAccountListIntegration` in `internal/cli/account_integration_test.go` to assert the "Converted" column header

## 3. Verify

- [x] 3.1 Run `make test` to ensure all tests pass
- [x] 3.2 Run `make coverage-check` to ensure 100% coverage is maintained
- [x] 3.3 Run `make lint` to ensure no linting issues
