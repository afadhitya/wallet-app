## 1. Create shared sub-package

- [x] 1.1 Create `internal/service/shared/` directory
- [x] 1.2 Create `shared.go` with `NotFoundError`, `ValidationError`, sentinel errors (`ErrNotFound`, `ErrDuplicateName`, `ErrInvalidAmount`, `ErrMissingField`)
- [x] 1.3 Create `resolvers.go` with `ResolveCategory(q, name)`, `ResolveAccount(q, name)`, `ResolveTag(q, name)` functions using `gen.Querier`
- [x] 1.4 Create `currency.go` with `GetBaseCurrency()`, `Convert(amount, currency)`, and `LoadRates`/`SaveRates` package-level vars
- [x] 1.5 Create `dates.go` with `ParseDate(input)`, `ParseMonth(input)` pure functions
- [x] 1.6 Update `service.go`: add thin delegation methods calling `shared.*` functions, remove old error type/helper definitions

## 2. Update service.go for manager embedding

- [x] 2.1 Add `*transaction.Manager`, `*plannedpayment.Manager`, `*budget.Manager`, `*report.Manager` as embedded fields in `Service` struct
- [x] 2.2 Update `New` and `NewWithQuerier` to wire all four managers via their constructors
- [x] 2.3 Update `internal/cli/helpers.go` and test files: replace `service.NotFoundError` / `service.ErrNotFound` imports with `shared.NotFoundError` / `shared.ErrNotFound`
- [x] 2.4 Verify `go build ./...` compiles (will fail until sub-packages exist — expected)

## 3. Extract transaction sub-package

- [x] 3.1 Create `internal/service/transaction/manager.go` with `Manager` struct and `NewManager(q gen.Querier) *Manager`
- [x] 3.2 Create `internal/service/transaction/types.go` with all params/result structs (AddExpenseParams, AddIncomeParams, AddTransferParams, TransferResult, AdjustBalanceParams, AdjustBalanceResult, EditTransactionParams, ListTransactionsParams, ListTransactionsResult, etc.)
- [x] 3.3 Create `internal/service/transaction/expense.go` with `AddExpense` method
- [x] 3.4 Create `internal/service/transaction/income.go` with `AddIncome` method
- [x] 3.5 Create `internal/service/transaction/transfer.go` with `AddTransfer` method and `TransferResult`
- [x] 3.6 Create `internal/service/transaction/adjustment.go` with `AdjustBalance` method
- [x] 3.7 Create `internal/service/transaction/edit.go` with `EditTransaction` and `RemoveTransaction` methods
- [x] 3.8 Create `internal/service/transaction/list.go` with `ListTransactions`, filter helpers, and `resolveBaseFields`
- [x] 3.9 Migrate cross-calls: `s.ResolveCategory` → `shared.ResolveCategory(m.q, ...)`, `s.ctx()` → `context.Background()`, `s.AddTransactionTag` → `m.q.AddTransactionTag(ctx.Background(), ...)`, `s.recalculateBalance` → recalcBalance helper
- [x] 3.10 Move transaction tests from `service_test.go` to `internal/service/transaction/*_test.go`
- [x] 3.11 Delete `internal/service/transaction.go`

## 4. Extract planned payment sub-package

- [x] 4.1 Create `internal/service/plannedpayment/manager.go` with `Manager` struct and `NewManager(q gen.Querier) *Manager`
- [x] 4.2 Create `internal/service/plannedpayment/types.go` with all params/result structs (CreatePlannedPaymentParams, EditPlannedPaymentParams, PayPlannedPaymentParams, etc.)
- [x] 4.3 Create `internal/service/plannedpayment/create_edit.go` with `CreatePlannedPayment` and `EditPlannedPayment`
- [x] 4.4 Create `internal/service/plannedpayment/pay_skip.go` with `PayPlannedPayment` and `SkipPlannedPayment`
- [x] 4.5 Create `internal/service/plannedpayment/states.go` with `PausePlannedPayment`, `ResumePlannedPayment`, `DeletePlannedPayment`
- [x] 4.6 Create `internal/service/plannedpayment/rrule.go` with calc helpers, RRULE parser, and `validateRRULE`
- [x] 4.7 Migrate cross-calls: replace service method calls with `shared.*` / `m.q.*` / Manager method calls
- [x] 4.8 Move planned payment tests from `service_test.go` to `internal/service/plannedpayment/*_test.go`
- [x] 4.9 Delete `internal/service/planned_payment.go`

## 5. Extract budget sub-package

- [x] 5.1 Create `internal/service/budget/manager.go` with `Manager` struct and `NewManager(q gen.Querier) *Manager`
- [x] 5.2 Create `internal/service/budget/types.go` with all params/result structs
- [x] 5.3 Create `internal/service/budget/create.go` with `SetBudget` and `updateExistingBudget`
- [x] 5.4 Create `internal/service/budget/check.go` with `CheckBudgets`, rollover, `resolveBudget`, and `buildCheckResult`
- [x] 5.5 Create `internal/service/budget/spending.go` with `calculateSpending`
- [x] 5.6 Create `internal/service/budget/edit.go` with `EditBudget` and `RemoveBudget`
- [x] 5.7 Migrate cross-calls: replace service method calls with `shared.*` / `m.q.*` / Manager method calls
- [x] 5.8 Move budget tests from `service_test.go` to `internal/service/budget/*_test.go`
- [x] 5.9 Delete `internal/service/budget.go`

## 6. Extract report sub-package

- [x] 6.1 Create `internal/service/report/manager.go` with `Manager` struct and `NewManager(q gen.Querier) *Manager`
- [x] 6.2 Create `internal/service/report/types.go` with `ReportParams`, `ReportFilters`, `ReportResult`, and all `*Row` types
- [x] 6.3 Create `internal/service/report/breakdown.go` with `generateMonthlySummary`, `generateCategoryBreakdown`, and related report generators
- [x] 6.4 Create `internal/service/report/export.go` with `GenerateExportRows` and `DefaultExportFilename`
- [x] 6.5 Migrate cross-calls: `GenerateReport` dispatcher uses Manager methods; replace service method calls
- [x] 6.6 Move report tests from `service_test.go` to `internal/service/report/*_test.go`
- [x] 6.7 Delete `internal/service/report.go`

## 7. Cleanup and verification

- [x] 7.1 Clean up `service.go`: remove any orphaned methods or unused imports
- [x] 7.2 Run `go build ./...` to verify compilation
- [x] 7.3 Run `go test ./internal/...` to verify all tests pass
- [x] 7.4 Run `make coverage-check` to verify 100% coverage maintained
- [x] 7.5 Update `coverignore.txt` if needed for new file paths
- [x] 7.6 Run linter and typecheck to verify code quality
