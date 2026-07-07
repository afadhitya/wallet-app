## 1. Fix service test infrastructure

- [x] 1.1 Add `SetTestRateConfig` call to `setupService()` in `internal/service/service_test.go` with base currency `"IDR"` and empty rates map, plus `t.Cleanup(ResetTestRateConfig)` for test isolation

## 2. Verification

- [x] 2.1 Run `go test ./internal/service/... -v` to confirm all tests pass
- [x] 2.2 Run `golangci-lint run ./...` to confirm no lint errors
- [x] 2.3 Run `make coverage-check` to confirm 100% coverage maintained
