package service

import (
	"database/sql"
	"testing"
	"time"
)

func setupServiceForForecast(t *testing.T) *Service {
	t.Helper()
	return setupService(t)
}

func TestForecastBalance_DefaultOneMonth(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	if _, err := svc.CreateAccount("BCA", "checking", "IDR"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Netflix",
		Amount:     149000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	})
	if err != nil {
		t.Fatalf("create PP: %v", err)
	}

	result, err := svc.ForecastBalance(1, "")
	if err != nil {
		t.Fatalf("ForecastBalance: %v", err)
	}
	if result.Months != 1 {
		t.Errorf("expected months 1, got %d", result.Months)
	}
	if len(result.MonthlyBalances) != 1 {
		t.Fatalf("expected 1 monthly balance, got %d", len(result.MonthlyBalances))
	}

	mb := result.MonthlyBalances[0]
	if mb.StartBalance < 0 {
		t.Errorf("expected non-negative start balance, got %d", mb.StartBalance)
	}
	if mb.ProjectedExpenses <= 0 {
		t.Errorf("expected projected expenses > 0, got %d", mb.ProjectedExpenses)
	}
	if mb.ProjectedIncome != 0 {
		t.Errorf("expected projected income 0, got %d", mb.ProjectedIncome)
	}
	if mb.EndingBalance != mb.StartBalance-mb.ProjectedExpenses {
		t.Errorf("ending balance mismatch: %d != %d - %d", mb.EndingBalance, mb.StartBalance, mb.ProjectedExpenses)
	}
	if len(result.PlannedPayments) == 0 {
		t.Error("expected at least one planned payment occurrence")
	}
}

func TestForecastBalance_MultiMonthRecurrence(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	if _, err := svc.CreateAccount("BCA", "checking", "IDR"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Netflix",
		Amount:     149000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	})
	if err != nil {
		t.Fatalf("create PP: %v", err)
	}

	result, err := svc.ForecastBalance(3, "")
	if err != nil {
		t.Fatalf("ForecastBalance: %v", err)
	}
	if result.Months != 3 {
		t.Errorf("expected months 3, got %d", result.Months)
	}
	if len(result.MonthlyBalances) != 3 {
		t.Fatalf("expected 3 monthly balances, got %d", len(result.MonthlyBalances))
	}

	if result.MonthlyBalances[1].EndingBalance != result.MonthlyBalances[2].StartBalance {
		t.Errorf("monthly balance continuity: month 2 end (%d) != month 3 start (%d)",
			result.MonthlyBalances[1].EndingBalance,
			result.MonthlyBalances[2].StartBalance)
	}
}

func TestForecastBalance_AccountFiltering(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	bca, err := svc.CreateAccount("BCA", "checking", "IDR")
	if err != nil {
		t.Fatalf("create BCA: %v", err)
	}
	mandiri, err := svc.CreateAccount("Mandiri", "savings", "IDR")
	if err != nil {
		t.Fatalf("create Mandiri: %v", err)
	}

	if _, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "BCA Bill",
		Amount:     100000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	}); err != nil {
		t.Fatalf("create BCA PP: %v", err)
	}
	if _, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Mandiri Bill",
		Amount:     50000,
		Account:    "Mandiri",
		Category:   "Coffee & Snacks",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	}); err != nil {
		t.Fatalf("create Mandiri PP: %v", err)
	}

	result, err := svc.ForecastBalance(1, "BCA")
	if err != nil {
		t.Fatalf("ForecastBalance BCA: %v", err)
	}
	if result.AccountName != "BCA" {
		t.Errorf("expected account BCA, got %s", result.AccountName)
	}
	if result.AccountID != bca.ID {
		t.Errorf("expected account ID %d, got %d", bca.ID, result.AccountID)
	}

	allResult, err := svc.ForecastBalance(1, "")
	if err != nil {
		t.Fatalf("ForecastBalance all: %v", err)
	}
	if len(allResult.PlannedPayments) <= len(result.PlannedPayments) {
		t.Error("expected all-account forecast to have more planned payments than single-account")
	}
	_ = mandiri
}

func TestForecastBalance_InvalidMonths(t *testing.T) {
	svc := setupServiceForForecast(t)

	_, err := svc.ForecastBalance(0, "")
	if err == nil {
		t.Fatal("expected error for zero months")
	}
	ve, ok := err.(*ValidationError)
	if !ok || ve.Field != "months" {
		t.Errorf("expected months validation error, got %v", err)
	}

	_, err = svc.ForecastBalance(-1, "")
	if err == nil {
		t.Fatal("expected error for negative months")
	}
}

func TestForecastBalance_NoPlannedPayments(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	if _, err := svc.CreateAccount("BCA", "checking", "IDR"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.ForecastBalance(1, "")
	if err != nil {
		t.Fatalf("ForecastBalance: %v", err)
	}
	if len(result.PlannedPayments) != 0 {
		t.Errorf("expected 0 planned payments, got %d", len(result.PlannedPayments))
	}
	if len(result.MonthlyBalances) != 1 {
		t.Errorf("expected 1 monthly balance, got %d", len(result.MonthlyBalances))
	}
	if result.MonthlyBalances[0].StartBalance != result.MonthlyBalances[0].EndingBalance {
		t.Error("expected flat balance with no payments")
	}
}

func TestForecastBalance_NegativeBalanceWarning(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	_, err := svc.CreateAccount("BCA", "checking", "IDR")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Rent",
		Amount:     5000000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	})
	if err != nil {
		t.Fatalf("create PP: %v", err)
	}

	result, err := svc.ForecastBalance(1, "")
	if err != nil {
		t.Fatalf("ForecastBalance: %v", err)
	}

	mb := result.MonthlyBalances[0]
	if !mb.IsNegative {
		t.Error("expected IsNegative to be true")
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warnings for negative balance")
	}
}

func TestForecastBalance_UnknownAccount(t *testing.T) {
	svc := setupServiceForForecast(t)

	_, err := svc.ForecastBalance(1, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown account")
	}
	_, ok := err.(*NotFoundError)
	if !ok {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestForecastBills_DefaultTwoMonths(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	if _, err := svc.CreateAccount("BCA", "checking", "IDR"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Netflix",
		Amount:     149000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	})
	if err != nil {
		t.Fatalf("create PP: %v", err)
	}

	result, err := svc.ForecastBills(2)
	if err != nil {
		t.Fatalf("ForecastBills: %v", err)
	}
	if result.Months != 2 {
		t.Errorf("expected months 2, got %d", result.Months)
	}
	if len(result.Bills) == 0 {
		t.Error("expected at least one bill")
	}
	if result.TotalAmount <= 0 {
		t.Error("expected positive total amount")
	}
	if result.Bills[0].RunningTotal <= 0 {
		t.Error("expected positive running total")
	}
}

func TestForecastBills_NoBills(t *testing.T) {
	svc := setupServiceForForecast(t)
	originalToday := todayFunc
	todayFunc = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { todayFunc = originalToday })

	if _, err := svc.CreateAccount("BCA", "checking", "IDR"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.ForecastBills(2)
	if err != nil {
		t.Fatalf("ForecastBills: %v", err)
	}
	if len(result.Bills) != 0 {
		t.Errorf("expected 0 bills, got %d", len(result.Bills))
	}
	if result.TotalAmount != 0 {
		t.Errorf("expected total amount 0, got %d", result.TotalAmount)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected empty-state warning")
	}
}

func TestForecastBills_InvalidMonths(t *testing.T) {
	svc := setupServiceForForecast(t)

	_, err := svc.ForecastBills(0)
	if err == nil {
		t.Fatal("expected error for zero months")
	}
}

func TestExpandOccurrences_OneTimeInRange(t *testing.T) {
	seed := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	occs := expandOccurrences(seed, "none", sql.NullString{}, start, end)
	if len(occs) != 1 {
		t.Errorf("expected 1 occurrence, got %d", len(occs))
	}
}

func TestExpandOccurrences_OneTimeOutOfRange(t *testing.T) {
	seed := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	occs := expandOccurrences(seed, "none", sql.NullString{}, start, end)
	if len(occs) != 0 {
		t.Errorf("expected 0 occurrences, got %d", len(occs))
	}
}

func TestExpandOccurrences_MonthlyMultiple(t *testing.T) {
	seed := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	occs := expandOccurrences(seed, "monthly", sql.NullString{}, start, end)
	if len(occs) != 3 {
		t.Errorf("expected 3 occurrences, got %d: %v", len(occs), occs)
	}
}

func TestExpandOccurrences_DailyMultiple(t *testing.T) {
	seed := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)

	occs := expandOccurrences(seed, "daily", sql.NullString{}, start, end)
	if len(occs) != 5 {
		t.Errorf("expected 5 occurrences, got %d", len(occs))
	}
}

func TestMonthsBetween(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		date   time.Time
		expect int
	}{
		{time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), 2},
		{time.Date(2027, 7, 1, 0, 0, 0, 0, time.UTC), 12},
	}

	for _, c := range cases {
		got := monthsBetween(start, c.date)
		if got != c.expect {
			t.Errorf("monthsBetween(start, %v) = %d, want %d", c.date, got, c.expect)
		}
	}
}
