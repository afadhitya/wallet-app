## Why

Users with accounts in multiple currencies currently see only raw balances per account in the list table. They have to mentally convert each non-base-currency balance to understand its value in the base currency. Showing the converted balance alongside the raw balance gives immediate clarity into the relative value of each account.

## What Changes

- Add a "Converted" column to the `wallet account list` table that shows each non-base-currency account balance converted to the base currency
- Base-currency accounts show nothing (or `-`) in the converted column since their raw balance is already in base currency
- Accounts with missing exchange rates show `-` (excluded) in the converted column, consistent with how they are excluded from the total
- The existing "Balance" column remains unchanged, showing the raw balance in the account's own currency
- **BREAKING**: The table column layout changes from `ID | Name | Type | Currency | Balance | Status` to `ID | Name | Type | Currency | Balance | Converted | Status`

## Capabilities

### New Capabilities

- `account-list-converted-balance`: Display a per-row converted balance in base currency for each account in the account list table

### Modified Capabilities

<!-- No existing spec requirements change. The new column is additive. -->

## Impact

- **`internal/cli/account.go`**: `runAccountList` function — add converted balance column to table output, format converted amount per row
- **`internal/cli/format.go`**: May need a new `formatAmountWithCurrency(amount int64, currency string)` or a parameterized variant to display amounts with appropriate currency symbols (not hardcoded `Rp`)
- **`internal/cli/account_test.go`**: Update table output assertions to include the new "Converted" column; add test cases for base-currency, non-base-currency, and missing-rate accounts
- **`internal/cli/account_integration_test.go`**: Update integration test assertions for new column
- **Existing specs**: No requirement-level changes — this is a new capability layered on top
