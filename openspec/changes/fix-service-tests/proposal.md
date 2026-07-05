## Why

Service tests fail because `setupService()` does not configure an in-memory rate config, causing `AddExpense`, `AddIncome`, and any test that creates transactions through them to try reading the real filesystem for `~/.config/wallet/rates.toml`. This breaks CI where the file does not exist.

## What Changes

- Add `SetTestRateConfig` call to `setupService()` so all service tests get a default IDR-only in-memory rate config
- Ensure `ResetTestRateConfig` runs via `t.Cleanup` for test isolation

## Capabilities

### New Capabilities

<!-- No new capabilities — this is a test infrastructure fix. -->

### Modified Capabilities

<!-- No spec-level requirement changes. -->

## Impact

- `internal/service/service_test.go` — `setupService()` helper
- No API, dependency, or system changes
