package service

import (
	"testing"

	"github.com/afadhitya/wallet-app/internal/testdb"
)

func TestGenerateReportBaseCurrencyOnly(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-02",
	})
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 200000, Description: "Freelance", Category: "Freelance",
		Account: "BCA", Date: "2026-07-03",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800, "EUR": 17200},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("Wise USD", "checking", "USD")
	_, _ = svc.CreateAccount("Revolut EUR", "checking", "EUR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 10, Description: "AWS", Category: "Subscriptions",
		Account: "Wise USD", Date: "2026-07-02",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 20, Description: "Hotel", Category: "Travel",
		Account: "Revolut EUR", Date: "2026-07-03",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AdjustBalance(AdjustBalanceParams{
		Account: "BCA", Target: 0, Description: "Adjust",
	})
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 100000, Description: "Salary", Category: "Salary",
		Account: "BCA", Date: "2026-07-02",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 500000, Description: "Salary", Category: "Salary",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddTransfer(CreateTransferParams{
		Amount: 100000, FromAccount: "BCA", ToAccount: "GoPay",
		Date: "2026-07-02",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-03",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	result, err := svc.GenerateReport(ReportParams{Month: "january"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA",
	})

	result, err := svc.GenerateReport(ReportParams{})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportAccountFilter(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("WiseUSD", "checking", "USD")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 10, Description: "AWS", Category: "Subscriptions",
		Account: "WiseUSD", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{AccountName: "BCA"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000 for BCA only, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportDateFilters(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-06-15",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("Wise USD", "checking", "USD")
	_, _ = svc.CreateAccount("Revolut EUR", "checking", "EUR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 10, Description: "AWS", Category: "Subscriptions",
		Account: "Wise USD", Date: "2026-07-02",
	})
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 100, Description: "Client payment", Category: "Freelance",
		Account: "Wise USD", Date: "2026-07-03",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 20, Description: "Hotel", Category: "Travel",
		Account: "Revolut EUR", Date: "2026-07-04",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july", By: "account"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july", By: "category"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	workTag, _ := svc.CreateTag("work")
	foodTag, _ := svc.CreateTag("food")

	txn1, _ := svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = svc.AddTransactionTag(txn1.Transaction.ID, workTag.ID)

	txn2, _ := svc.AddExpense(CreateExpenseParams{
		Amount: 30000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-02",
	})
	_ = svc.AddTransactionTag(txn2.Transaction.ID, foodTag.ID)

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 20000, Description: "Snack", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-03",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july", By: "tag"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	workTag, _ := svc.CreateTag("work")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = svc.AddTransactionTag(txn.Transaction.ID, workTag.ID)

	result, err := svc.GenerateReport(ReportParams{Month: "july", By: "tag"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if len(result.ByTag) != 1 {
		t.Errorf("expected 1 tag row, got %d", len(result.ByTag))
	}
}

func TestGenerateReportAccountFilterWithByCategory(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "GoPay", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "july", By: "category", AccountName: "BCA"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-10",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})

	result, err := svc.GenerateReport(ReportParams{Month: "2026-07"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.ExpenseTotal != 50000 {
		t.Errorf("expected expense 50000, got %d", result.ExpenseTotal)
	}
}

func TestGenerateReportInvalidMonth(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, err := svc.GenerateReport(ReportParams{Month: "not-a-month"})
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestGenerateReportInvalidBy(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, err := svc.GenerateReport(ReportParams{Month: "july", By: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid --by")
	}
}

func TestGeneratePeriodLabel(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

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
			result, err := svc.GenerateReport(tt.params)
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	workTag, _ := svc.CreateTag("work")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = svc.AddTransactionTag(txn.Transaction.ID, workTag.ID)

	rows, err := svc.GenerateExportRows(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	tagA, _ := svc.CreateTag("work")
	tagB, _ := svc.CreateTag("lunch")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount: 50000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = svc.AddTransactionTag(txn.Transaction.ID, tagA.ID)
	_ = svc.AddTransactionTag(txn.Transaction.ID, tagB.ID)

	rows, err := svc.GenerateExportRows(ReportParams{Month: "july"})
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
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	filename, err := svc.DefaultExportFilename(ReportParams{Month: "2026-07"})
	if err != nil {
		t.Fatalf("DefaultExportFilename: %v", err)
	}

	if filename != "wallet-report-2026-07.csv" {
		t.Errorf("expected 'wallet-report-2026-07.csv', got '%s'", filename)
	}
}

func TestGenerateReportDateRangeValidation(t *testing.T) {
	svc := New(testdb.Open(t, testLogger()), testLogger())
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	_, err := svc.GenerateReport(ReportParams{DateFrom: "2026-01-01"})
	if err == nil {
		t.Fatal("expected error when only --from is provided")
	}

	_, err = svc.GenerateReport(ReportParams{DateTo: "2026-01-31"})
	if err == nil {
		t.Fatal("expected error when only --to is provided")
	}
}
