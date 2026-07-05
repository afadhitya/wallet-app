## Context

`setupService()` in `internal/service/service_test.go` creates a `*Service` from an in-memory SQLite database but does not configure an in-memory rate config. The service's `AddExpense` and `AddIncome` methods call `resolveBaseFields()` which calls `GetBaseCurrency()` which calls `svcLoadRates()` (defaults to `config.LoadRates` reading `~/.config/wallet/rates.toml`). This causes any test using `setupService()` for transaction creation to fail in CI where the file doesn't exist.

Other test files (`currency_test.go`, `budget_test.go`, `report_test.go`) already handle this by calling `SetTestRateConfig` in their own setup helpers. The `setupServiceWithMultiCurrency()` helper at the bottom of `service_test.go` also correctly calls `SetTestRateConfig`.

## Goals / Non-Goals

**Goals:**
- Fix `setupService()` so tests pass without filesystem rate config
- Use `t.Cleanup(ResetTestRateConfig)` for test isolation

**Non-Goals:**
- Changing `AddExpense`/`AddIncome` to make `GetBaseCurrency()` optional
- Modifying `resolveBaseFields()` behavior
- Adding rate config to tests that don't need it (e.g., `TestCreateAccount`)

## Decisions

**Decision: Add `SetTestRateConfig` + `t.Cleanup(ResetTestRateConfig)` to `setupService()`**

Using existing `SetTestRateConfig`/`ResetTestRateConfig` from `internal/service/currency.go`:
- `SetTestRateConfig(TestRateConfig{BaseCurrency: "IDR", Rates: map[string]int64{}})` — all tests operate in IDR
- `t.Cleanup(ResetTestRateConfig)` — restores filesystem loader between tests, prevents cross-test pollution

This is safe for `setupServiceWithMultiCurrency()` which calls `setupService()` then immediately overrides the rate config with its own `SetTestRateConfig` call — the second call just replaces the in-memory config.

## Risks / Trade-offs

- [Risk: tests accidentally rely on global state] → Mitigation: `t.Cleanup` ensures each test resets. The pattern is already used in `currency_test.go` and `budget_test.go`.
- [Trade-off: rate config init overhead per test] → Negligible; in-memory map allocation.
