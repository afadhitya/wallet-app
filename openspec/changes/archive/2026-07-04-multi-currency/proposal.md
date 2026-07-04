## Why

Wallet transactions can already belong to accounts with different currencies, but the application does not yet provide a user-facing way to manage exchange rates, lock conversion values at transaction time, or report mixed-currency activity in a single base currency. Multi-currency support is needed so travel, foreign-bank, and cross-border spending remains accurate offline and does not depend on live exchange-rate services.

## What Changes

- Add manual exchange-rate configuration through `~/.config/wallet/rates.toml`, with a configurable base currency that defaults to `IDR`.
- Add `wallet rate list`, `wallet rate set`, `wallet rate add`, and `wallet rate rm` commands, including JSON output for listing rates.
- Convert foreign-currency income and expense transactions at creation time using the account currency and configured rate.
- Store the original transaction amount/currency and locked base-currency equivalent so later rate changes do not rewrite history.
- Update transaction listing and reporting behavior so mixed-currency data is displayed with both original and base-currency amounts where useful.
- Return clear validation errors for missing rates, invalid rates, and missing rate configuration.

## Capabilities

### New Capabilities
- `multi-currency`: Rate configuration, conversion service behavior, rate-management CLI commands, transaction-time conversion, and mixed-currency display/reporting behavior.

### Modified Capabilities
- `core-crud`: Transaction entry, listing, initialization, and service behavior must account for account currency, locked base amounts, and rate configuration.
- `wallet-data-model`: Clarifies how the existing transaction currency/base amount fields and exchange-rate storage are used for locked transaction conversions.

## Impact

- Affected CLI commands: `wallet init`, `wallet rate *`, `wallet add expense`, `wallet add income`, `wallet list`, and reporting commands that aggregate transaction totals.
- Affected layers: configuration loading/writing, service validation, transaction persistence, account balance updates, output formatting, JSON rendering, and tests.
- No live exchange-rate dependency is introduced; rates are manually managed in local TOML config.
- Existing transactions are not retroactively changed when rates are added, updated, or removed.
