package service

import (
	"testing"

	"github.com/afadhitya/wallet-app/internal/testdb"
)

func TestGenerateReportBaseCurrencyOnly(t *testing.T) {
	svc := New(testdb.Open(t))
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
	if result.TotalIncome != 200000 {
		t.Errorf("expected total income 200000, got %d", result.TotalIncome)
	}
	if result.TotalExpense != 85000 {
		t.Errorf("expected total expense 85000, got %d", result.TotalExpense)
	}
	if result.Net != 115000 {
		t.Errorf("expected net 115000, got %d", result.Net)
	}
	if len(result.ByCategory) < 3 {
		t.Errorf("expected at least 3 categories, got %d", len(result.ByCategory))
	}
}

func TestGenerateReportMixedCurrency(t *testing.T) {
	svc := New(testdb.Open(t))
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
	if result.TotalExpense != expectedExpense {
		t.Errorf("expected total expense %d, got %d", expectedExpense, result.TotalExpense)
	}
	if result.TotalIncome != 0 {
		t.Errorf("expected total income 0, got %d", result.TotalIncome)
	}
}

func TestGenerateReportExcludesAdjustment(t *testing.T) {
	svc := New(testdb.Open(t))
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

	if result.TotalIncome != 100000 {
		t.Errorf("expected income 100000 (excluding adjustment), got %d", result.TotalIncome)
	}
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000, got %d", result.TotalExpense)
	}
}

func TestGenerateReportExcludesTransfer(t *testing.T) {
	svc := New(testdb.Open(t))
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

	if result.TotalIncome != 500000 {
		t.Errorf("expected income 500000 (excluding transfer), got %d", result.TotalIncome)
	}
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000 (excluding transfer), got %d", result.TotalExpense)
	}
}

func TestGenerateReportNoTransactions(t *testing.T) {
	svc := New(testdb.Open(t))
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	defer ResetTestRateConfig()

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	if result.TotalIncome != 0 {
		t.Errorf("expected zero income, got %d", result.TotalIncome)
	}
	if result.TotalExpense != 0 {
		t.Errorf("expected zero expense, got %d", result.TotalExpense)
	}
}

func TestGenerateReportNoFilterDefaultsToCurrentMonth(t *testing.T) {
	svc := New(testdb.Open(t))
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
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000, got %d", result.TotalExpense)
	}
}

func TestFetchReportTransactionsAccountFilter(t *testing.T) {
	svc := New(testdb.Open(t))
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
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000 for BCA only, got %d", result.TotalExpense)
	}
}

func TestFetchReportTransactionsCategoryFilter(t *testing.T) {
	svc := New(testdb.Open(t))
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

	result, err := svc.GenerateReport(ReportParams{CategoryName: "Coffee & Snacks"})
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000 for Coffee & Snacks only, got %d", result.TotalExpense)
	}
}

func TestFetchReportTransactionsDateFilters(t *testing.T) {
	svc := New(testdb.Open(t))
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
	if result.TotalExpense != 50000 {
		t.Errorf("expected expense 50000 for June only, got %d", result.TotalExpense)
	}
}

func TestGenerateReportByAccount(t *testing.T) {
	svc := New(testdb.Open(t))
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

	result, err := svc.GenerateReport(ReportParams{Month: "july"})
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
