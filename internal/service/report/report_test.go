package report

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
	"github.com/afadhitya/wallet-app/internal/testdb"
)

func setupReportManager(t *testing.T) *ReportManager {
	t.Helper()
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	return NewReportManager(gen.New(testdb.Open(t)))
}

func mustCreateAccount(t *testing.T, m *ReportManager, name, typ, currency string) *gen.Account {
	t.Helper()
	acc, err := m.q.CreateAccount(context.Background(), gen.CreateAccountParams{
		Name: name, Type: typ, Currency: currency,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return acc
}

func resolveBaseFields(currency string, amount int64) (sql.NullInt64, sql.NullString, error) {
	baseCurrency, err := shared.GetBaseCurrency()
	if err != nil {
		return sql.NullInt64{}, sql.NullString{}, err
	}
	if currency == baseCurrency {
		return sql.NullInt64{}, sql.NullString{}, nil
	}
	converted, err := shared.Convert(amount, currency)
	if err != nil {
		return sql.NullInt64{}, sql.NullString{}, err
	}
	return sql.NullInt64{Int64: converted, Valid: true}, sql.NullString{String: baseCurrency, Valid: true}, nil
}

func mustCreateTransaction(t *testing.T, m *ReportManager, txnType, accountName, categoryName string, amount int64, description, date string) *gen.Transaction {
	t.Helper()
	account, err := shared.ResolveAccount(m.q, accountName)
	if err != nil {
		t.Fatalf("resolve account %q: %v", accountName, err)
	}
	var catID sql.NullInt64
	if categoryName != "" {
		cat, err := shared.ResolveCategory(m.q, categoryName)
		if err != nil {
			t.Fatalf("resolve category %q: %v", categoryName, err)
		}
		catID = sql.NullInt64{Int64: cat.ID, Valid: true}
	}
	baseAmount, baseCurrency, err := resolveBaseFields(account.Currency, amount)
	if err != nil {
		t.Fatalf("resolve base fields: %v", err)
	}
	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}
	txn, err := m.q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID:    account.ID,
		CategoryID:   catID,
		Type:         txnType,
		Amount:       amount,
		Currency:     account.Currency,
		Description:  desc,
		Date:         date,
		BaseAmount:   baseAmount,
		BaseCurrency: baseCurrency,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	return txn
}

func mustAddExpense(t *testing.T, m *ReportManager, accountName, categoryName string, amount int64, description, date string) *gen.Transaction {
	t.Helper()
	return mustCreateTransaction(t, m, "expense", accountName, categoryName, amount, description, date)
}

func mustAddIncome(t *testing.T, m *ReportManager, accountName, categoryName string, amount int64, description, date string) *gen.Transaction {
	t.Helper()
	return mustCreateTransaction(t, m, "income", accountName, categoryName, amount, description, date)
}

func mustAddTransfer(t *testing.T, m *ReportManager, fromAccountName, toAccountName string, amount int64, date string) *gen.Transaction {
	t.Helper()
	fromAccount, err := shared.ResolveAccount(m.q, fromAccountName)
	if err != nil {
		t.Fatalf("resolve from account: %v", err)
	}
	toAccount, err := shared.ResolveAccount(m.q, toAccountName)
	if err != nil {
		t.Fatalf("resolve to account: %v", err)
	}
	transferToID := sql.NullInt64{Int64: toAccount.ID, Valid: true}
	txn, err := m.q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID:    fromAccount.ID,
		CategoryID:   sql.NullInt64{},
		Type:         "transfer",
		Amount:       amount,
		Currency:     "IDR",
		TransferToID: transferToID,
		Date:         date,
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	return txn
}

func mustAddAdjustment(t *testing.T, m *ReportManager, accountName string, amount int64, description string) *gen.Transaction {
	t.Helper()
	account, err := shared.ResolveAccount(m.q, accountName)
	if err != nil {
		t.Fatalf("resolve account: %v", err)
	}
	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}
	txn, err := m.q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID:   account.ID,
		CategoryID:  sql.NullInt64{},
		Type:        "adjustment",
		Amount:      amount,
		Currency:    "IDR",
		Description: desc,
		Date:        time.Now().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}
	return txn
}

func mustCreateTag(t *testing.T, m *ReportManager, name string) *gen.Tag {
	t.Helper()
	tag, err := m.q.CreateTag(context.Background(), name)
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	return tag
}

func mustAddTag(t *testing.T, m *ReportManager, txnID, tagID int64) {
	t.Helper()
	if err := m.q.AddTransactionTag(context.Background(), gen.AddTransactionTagParams{
		TransactionID: txnID,
		TagID:         tagID,
	}); err != nil {
		t.Fatalf("add tag: %v", err)
	}
}

func TestGenerateReportBaseCurrencyOnly(t *testing.T) {
	m := setupReportManager(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "BCA", "Restaurant", 35000, "Lunch", "2026-07-02")
	mustAddIncome(t, m, "BCA", "Freelance", 200000, "Freelance", "2026-07-03")

	result, err := m.GenerateReport(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.BaseCurrency != "IDR" {
		t.Errorf("expected IDR, got %s", result.BaseCurrency)
	}
	if result.IncomeTotal != 200000 {
		t.Errorf("expected income_total 200000, got %d", result.IncomeTotal)
	}
	if result.ExpenseTotal != 85000 {
		t.Errorf("expected expense_total 85000, got %d", result.ExpenseTotal)
	}
	if result.Net != 115000 {
		t.Errorf("expected net 115000, got %d", result.Net)
	}
	if len(result.IncomeCategories)+len(result.ExpenseCategories) < 3 {
		t.Errorf("expected at least 3 categories, got %d", len(result.IncomeCategories)+len(result.ExpenseCategories))
	}
}

func TestGenerateReportMixedCurrency(t *testing.T) {
	m := setupReportManager(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800, "EUR": 17200},
	})

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustCreateAccount(t, m, "Wise USD", "checking", "USD")
	mustCreateAccount(t, m, "Revolut EUR", "checking", "EUR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "Wise USD", "Subscriptions", 10, "AWS", "2026-07-02")
	mustAddExpense(t, m, "Revolut EUR", "Travel", 20, "Hotel", "2026-07-03")

	result, err := m.GenerateReport(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	expectedExpense := int64(50000 + 10*15800 + 20*17200)
	if result.ExpenseTotal != expectedExpense {
		t.Errorf("expected expense_total %d, got %d", expectedExpense, result.ExpenseTotal)
	}
	if result.IncomeTotal != 0 {
		t.Errorf("expected income_total 0, got %d", result.IncomeTotal)
	}
}

func TestGenerateReportExcludesAdjustment(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddAdjustment(t, m, "BCA", 50000, "Adjust")
	mustAddIncome(t, m, "BCA", "Salary", 100000, "Salary", "2026-07-02")

	result, err := m.GenerateReport(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.IncomeTotal != 100000 {
		t.Errorf("expected income 100000 (excluding adjustment), got %d", result.IncomeTotal)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportExcludesTransfer(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	mustAddIncome(t, m, "BCA", "Salary", 500000, "Salary", "2026-07-01")
	mustAddTransfer(t, m, "BCA", "GoPay", 100000, "2026-07-02")
	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-03")

	result, err := m.GenerateReport(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.IncomeTotal != 500000 {
		t.Errorf("expected income 500000 (excluding transfer), got %d", result.IncomeTotal)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000 (excluding transfer), got %d", result.ExpenseTotal)
	}
	if result.TransferTotal != 100000 {
		t.Errorf("expected transfer_total 100000, got %d", result.TransferTotal)
	}
}

func TestGenerateReportNoTransactions(t *testing.T) {
	m := setupReportManager(t)

	result, err := m.GenerateReport(ReportParams{Month: "january"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.IncomeTotal != 0 {
		t.Errorf("expected zero income, got %d", result.IncomeTotal)
	}
	if result.ExpenseTotal != 0 {
		t.Errorf("expected zero expense, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportNoFilterDefaultsToCurrentMonth(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", time.Now().Format("2006-01-02"))

	result, err := m.GenerateReport(ReportParams{})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportAccountFilter(t *testing.T) {
	m := setupReportManager(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustCreateAccount(t, m, "WiseUSD", "checking", "USD")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "WiseUSD", "Subscriptions", 10, "AWS", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{AccountName: "BCA"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000 for BCA only, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportDateFilters(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-06-15")
	mustAddExpense(t, m, "BCA", "Restaurant", 35000, "Lunch", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{
		DateFrom: "2026-06-01",
		DateTo:   "2026-06-30",
	})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000 for June only, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportByAccount(t *testing.T) {
	m := setupReportManager(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustCreateAccount(t, m, "Wise USD", "checking", "USD")
	mustCreateAccount(t, m, "Revolut EUR", "checking", "EUR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "Wise USD", "Subscriptions", 10, "AWS", "2026-07-02")
	mustAddIncome(t, m, "Wise USD", "Freelance", 100, "Client payment", "2026-07-03")
	mustAddExpense(t, m, "Revolut EUR", "Travel", 20, "Hotel", "2026-07-04")

	result, err := m.GenerateReport(ReportParams{Month: "july", By: "account"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByAccount) != 3 {
		t.Errorf("expected 3 accounts, got %d", len(result.ByAccount))
	}

	for _, ab := range result.ByAccount {
		switch ab.AccountName {
		case "BCA":
			if ab.Expense != 50000 {
				t.Errorf("expected BCA expense 50000, got %d", ab.Expense)
			}
			if ab.Income != 0 {
				t.Errorf("expected BCA income 0, got %d", ab.Income)
			}
		case "Wise USD":
			if ab.Expense != 10*15800 {
				t.Errorf("expected Wise USD expense %d, got %d", 10*15800, ab.Expense)
			}
			if ab.Income != 100*15800 {
				t.Errorf("expected Wise USD income %d, got %d", 100*15800, ab.Income)
			}
		case "Revolut EUR":
			if ab.Expense != 20*17200 {
				t.Errorf("expected Revolut EUR expense %d, got %d", 20*17200, ab.Expense)
			}
		}
	}
}

func TestGenerateReportByCategory(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "BCA", "Restaurant", 35000, "Lunch", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{Month: "july", By: "category"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByCategory) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result.ByCategory))
	}
	if result.ExpenseTotal != 85000 {
		t.Errorf("expected expense_total 85000, got %d", result.ExpenseTotal)
	}

	for _, cat := range result.ByCategory {
		if cat.CategoryName == "Coffee & Snacks" {
			if cat.Percent < 58.0 || cat.Percent > 59.0 {
				t.Errorf("expected ~58.8%%, got %.1f%%", cat.Percent)
			}
			if cat.TransactionCount != 1 {
				t.Errorf("expected 1 txns, got %d", cat.TransactionCount)
			}
		}
	}
}

func TestGenerateReportByTag(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	workTag := mustCreateTag(t, m, "work")
	foodTag := mustCreateTag(t, m, "food")

	txn1 := mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddTag(t, m, txn1.ID, workTag.ID)

	txn2 := mustAddExpense(t, m, "BCA", "Restaurant", 30000, "Lunch", "2026-07-02")
	mustAddTag(t, m, txn2.ID, foodTag.ID)

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 20000, "Snack", "2026-07-03")

	result, err := m.GenerateReport(ReportParams{Month: "july", By: "tag"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByTag) != 3 {
		t.Errorf("expected 3 tag rows (work, food, untagged), got %d", len(result.ByTag))
	}
	if result.ExpenseTotal != 100000 {
		t.Errorf("expected expense_total 100000, got %d", result.ExpenseTotal)
	}

	foundUntagged := false
	for _, tr := range result.ByTag {
		if tr.TagName == "(untagged)" {
			foundUntagged = true
			if tr.Amount != 20000 {
				t.Errorf("expected untagged amount 20000, got %d", tr.Amount)
			}
		}
	}
	if !foundUntagged {
		t.Errorf("expected untagged row in tag breakdown")
	}
}

func TestGenerateReportByTagNoUntagged(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	workTag := mustCreateTag(t, m, "work")

	txn := mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddTag(t, m, txn.ID, workTag.ID)

	result, err := m.GenerateReport(ReportParams{Month: "july", By: "tag"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByTag) != 1 {
		t.Errorf("expected 1 tag row, got %d", len(result.ByTag))
	}
}

func TestGenerateReportAccountFilterWithByCategory(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")
	mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddExpense(t, m, "GoPay", "Restaurant", 35000, "Lunch", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{Month: "july", By: "category", AccountName: "BCA"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByCategory) != 1 {
		t.Errorf("expected 1 category for BCA, got %d", len(result.ByCategory))
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense_total 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportDateRangeOverridesMonth(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-10")
	mustAddExpense(t, m, "BCA", "Restaurant", 35000, "Lunch", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{
		Month:    "2026-07",
		DateFrom: "2026-07-10",
		DateTo:   "2026-07-20",
	})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000 for July 10-20 only, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportYYYYMM(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")

	result, err := m.GenerateReport(ReportParams{Month: "2026-07"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportInvalidMonth(t *testing.T) {
	m := setupReportManager(t)

	_, err := m.GenerateReport(ReportParams{Month: "not-a-month"})
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestGenerateReportInvalidBy(t *testing.T) {
	m := setupReportManager(t)

	_, err := m.GenerateReport(ReportParams{Month: "july", By: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid --by")
	}
}

func TestGeneratePeriodLabel(t *testing.T) {
	m := setupReportManager(t)

	tests := []struct {
		name     string
		params   ReportParams
		wantCont string
	}{
		{"month name", ReportParams{Month: "january"}, "January 2026"},
		{"YYYY-MM", ReportParams{Month: "2026-03"}, "March 2026"},
		{"date range", ReportParams{DateFrom: "2026-01-01", DateTo: "2026-01-31"}, "2026-01-01 to 2026-01-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GenerateReport(tt.params)
			if err != nil {
				t.Fatalf("GenerateReport: %v", err)
			}
			if result.Period != tt.wantCont {
				t.Errorf("expected period '%s', got '%s'", tt.wantCont, result.Period)
			}
		})
	}
}

func TestGenerateExportRows(t *testing.T) {
	m := setupReportManager(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	workTag := mustCreateTag(t, m, "work")

	txn := mustAddExpense(t, m, "BCA", "Coffee & Snacks", 50000, "Coffee", "2026-07-01")
	mustAddTag(t, m, txn.ID, workTag.ID)

	rows, err := m.GenerateExportRows(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateExportRows: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	r := rows[0]
	if r.Date != "2026-07-01" {
		t.Errorf("expected date 2026-07-01, got %s", r.Date)
	}
	if r.Type != "expense" {
		t.Errorf("expected type expense, got %s", r.Type)
	}
	if r.Amount != 50000 {
		t.Errorf("expected amount 50000, got %d", r.Amount)
	}
	if r.Currency != "IDR" {
		t.Errorf("expected currency IDR, got %s", r.Currency)
	}
	if r.Category != "Coffee & Snacks" {
		t.Errorf("expected category 'Coffee & Snacks', got '%s'", r.Category)
	}
	if r.Account != "BCA" {
		t.Errorf("expected account BCA, got %s", r.Account)
	}
	if r.Description != "Coffee" {
		t.Errorf("expected description Coffee, got %s", r.Description)
	}
	if r.Tags != "work" {
		t.Errorf("expected tags 'work', got '%s'", r.Tags)
	}
}

func TestGenerateExportRowsMultipleTags(t *testing.T) {
	m := setupReportManager(t)

	mustCreateAccount(t, m, "BCA", "checking", "IDR")

	tagA := mustCreateTag(t, m, "work")
	tagB := mustCreateTag(t, m, "lunch")

	txn := mustAddExpense(t, m, "BCA", "Restaurant", 50000, "Lunch", "2026-07-01")
	mustAddTag(t, m, txn.ID, tagA.ID)
	mustAddTag(t, m, txn.ID, tagB.ID)

	rows, err := m.GenerateExportRows(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateExportRows: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	if rows[0].Tags != "work,lunch" && rows[0].Tags != "lunch,work" {
		t.Errorf("expected tags 'work,lunch' or 'lunch,work', got '%s'", rows[0].Tags)
	}
}

func TestDefaultExportFilename(t *testing.T) {
	m := setupReportManager(t)

	filename, err := m.DefaultExportFilename(ReportParams{Month: "2026-07"})
	if err != nil {
		t.Fatalf("DefaultExportFilename: %v", err)
	}

	if filename != "wallet-report-2026-07.csv" {
		t.Errorf("expected 'wallet-report-2026-07.csv', got '%s'", filename)
	}
}

func TestGenerateReportDateRangeValidation(t *testing.T) {
	m := setupReportManager(t)

	_, err := m.GenerateReport(ReportParams{DateFrom: "2026-01-01"})
	if err == nil {
		t.Fatal("expected error when only --from is provided")
	}

	_, err = m.GenerateReport(ReportParams{DateTo: "2026-01-31"})
	if err == nil {
		t.Fatal("expected error when only --to is provided")
	}
}
