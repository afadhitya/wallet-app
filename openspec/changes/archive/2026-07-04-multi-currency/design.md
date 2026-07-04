## Context

The wallet schema already stores account currencies, transaction currency fields, optional locked base-currency amounts, and an `exchange_rates` table. Core CRUD commands currently treat entered amounts as account-local amounts and do not provide rate management, transaction-time conversion, or mixed-currency display semantics.

This change adds offline-first multi-currency behavior without adding network dependencies. Rates are controlled by the user in local configuration and are applied when transactions are created, so historical totals remain stable even when the user later edits rates.

## Goals / Non-Goals

**Goals:**
- Provide a local `rates.toml` configuration with base currency and rates expressed as base units per one foreign unit.
- Add a currency service that loads, validates, converts, and updates configured rates.
- Add rate-management CLI commands for listing, adding, setting, and removing rates.
- Convert foreign-currency income and expense transactions at creation time using the selected account currency.
- Display original and base-currency amounts in transaction lists and reports where mixed currencies are present.
- Keep existing transactions stable when rates change.

**Non-Goals:**
- Fetching live rates from external APIs.
- Revaluing historical transactions when rates change.
- Per-transaction manual rate override in this change.
- Cross-currency transfers with independent source and destination conversion semantics beyond the existing transfer behavior.

## Decisions

### Store rates in config, not as the primary source in the database

Rates will be read from `~/.config/wallet/rates.toml`, with `IDR` as the default base currency and a `[rates]` map where each value means `1 <currency> = <rate> <base>`. This keeps rate management manual, predictable, and offline-friendly.

Alternative considered: use the existing `exchange_rates` table as the authoritative source. That table remains useful for cached or future sourced rates, but making it authoritative now would require extra CRUD/query behavior and conflicts with the brainstorming decision to keep user-edited rates in config.

### Convert at transaction creation time

When an income or expense is recorded for an account whose currency differs from the base currency, the service will look up the configured rate and store `base_amount` and `base_currency` with the transaction. For base-currency accounts, those fields remain unset.

Alternative considered: convert dynamically during listing/reporting. Dynamic conversion would make historical reports change when rates are edited and would make reporting slower and less deterministic.

### Keep account balances in account currency

Account balance mutations continue to use the transaction amount in the account's own currency. Base amounts are for reporting and aggregation across currencies, not for mutating a foreign-currency account balance.

Alternative considered: store all account balances in base currency. That would make foreign account balances inaccurate from the user's perspective and would be a larger data-model change.

### Introduce a dedicated currency service

Currency behavior should live behind a service responsible for loading config, validating positive rates, converting integer minor-unit amounts, and persisting rate edits. Transaction services use this service instead of reading config directly.

Alternative considered: implement conversion directly in CLI handlers. That would duplicate validation and make service-level tests less complete.

## Risks / Trade-offs

- Integer rounding can differ from user expectations for currencies with different minor-unit conventions. Mitigation: convert with deterministic rounding and document that amounts are stored in the app's existing integer minor-unit representation.
- Manual rates can become stale. Mitigation: make the rate source explicit in commands and avoid retroactive updates.
- Missing rates will block foreign-currency transaction creation. Mitigation: return an actionable error that points to `wallet rate add <currency> <rate>`.
- Reporting mixed currencies can become visually noisy. Mitigation: show base totals as primary and original-currency context only where it improves understanding.

## Migration Plan

- Extend default initialization so missing rate configuration is created with base currency `IDR` and common sample rates only if the config file does not already exist.
- Existing transactions remain unchanged. Transactions without `base_amount` continue to be interpreted as base-currency transactions when their `currency` equals the configured base currency.
- Rollback is low risk because rate config is additive and historical transaction rows store locked base amounts independently from current config.

## Open Questions

- Should `wallet rate list` include rows from the `exchange_rates` table, or only the config-backed manual rates? This design uses config-backed rates only.
- Should users be able to override the conversion rate per transaction with a `--rate` flag? This remains out of scope for the first implementation.
