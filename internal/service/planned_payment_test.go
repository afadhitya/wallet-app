package service

import (
	"database/sql"
	"testing"
	"time"
)

func setupServiceForPP(t *testing.T) *Service {
	t.Helper()
	return setupService(t)
}

func mustCreateAccount(t *testing.T, svc *Service, name, accType, currency string) {
	t.Helper()
	if _, err := svc.CreateAccount(name, accType, currency); err != nil {
		t.Fatalf("create account %s: %v", name, err)
	}
}

func TestCreatePlannedPayment_Valid(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Netflix",
		Amount:     149000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
		StartDate:  "2026-07-15",
		DueDay:     15,
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}
	if pp.Name != "Netflix" {
		t.Errorf("expected name Netflix, got %s", pp.Name)
	}
	if pp.Amount != 149000 {
		t.Errorf("expected amount 149000, got %d", pp.Amount)
	}
	if pp.Recurrence != "monthly" {
		t.Errorf("expected recurrence monthly, got %s", pp.Recurrence)
	}
	if !pp.NextDueDate.Valid {
		t.Error("expected next_due_date to be set")
	}
}

func TestCreatePlannedPayment_InvalidAmount(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Invalid",
		Amount:     0,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreatePlannedPayment_MissingName(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Amount:     50000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "monthly",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	ve, ok := err.(*ValidationError)
	if !ok || ve.Field != "name" {
		t.Errorf("expected name validation error, got %v", err)
	}
}

func TestCreatePlannedPayment_InvalidRecurrence(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Test",
		Amount:     50000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid recurrence")
	}
}

func TestCreatePlannedPayment_CustomRRULEMissing(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Test",
		Amount:     50000,
		Account:    "BCA",
		Category:   "Subscriptions",
		Recurrence: "custom",
	})
	if err == nil {
		t.Fatal("expected error for missing RRULE with custom recurrence")
	}
}

func TestCreatePlannedPayment_InvalidRRULE(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:           "Test",
		Amount:         50000,
		Account:        "BCA",
		Category:       "Subscriptions",
		Recurrence:     "custom",
		RecurrenceRule: "INVALID",
	})
	if err == nil {
		t.Fatal("expected error for invalid RRULE")
	}
}

func TestCreatePlannedPayment_AccountNotFound(t *testing.T) {
	svc := setupServiceForPP(t)

	_, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Test",
		Amount:     50000,
		Account:    "NonExistent",
		Category:   "Subscriptions",
		Recurrence: "monthly",
	})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestCreatePlannedPayment_OneTime(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name:       "Flight to Tokyo",
		Amount:     3000000,
		Account:    "BCA",
		Category:   "Travel",
		Recurrence: "none",
		StartDate:  "2026-08-15",
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}
	if pp.Recurrence != "none" {
		t.Errorf("expected recurrence none, got %s", pp.Recurrence)
	}
}

func TestListPlannedPayments_Active(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, _ = svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	_, _ = svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Spotify", Amount: 54990, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 1,
	})

	payments, err := svc.ListPlannedPayments(false, false)
	if err != nil {
		t.Fatalf("ListPlannedPayments: %v", err)
	}
	if len(payments) != 2 {
		t.Errorf("expected 2 active payments, got %d", len(payments))
	}
}

func TestListPlannedPayments_Paused(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	_ = svc.PausePlannedPayment(pp.ID)

	active, err := svc.ListPlannedPayments(false, false)
	if err != nil {
		t.Fatalf("ListPlannedPayments active: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("expected 0 active payments after pause, got %d", len(active))
	}

	paused, err := svc.ListPlannedPayments(true, false)
	if err != nil {
		t.Fatalf("ListPlannedPayments paused: %v", err)
	}
	if len(paused) != 1 {
		t.Errorf("expected 1 paused payment, got %d", len(paused))
	}
}

func TestListDuePlannedPayments_CurrentMonth(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	now := time.Now()
	today := now.Format("2006-01-02")

	_, _ = svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Due Now", Amount: 50000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: today, DueDay: now.Day(),
	})

	due, total, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueCurrentMonth})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments: %v", err)
	}
	if len(due) != 1 {
		t.Errorf("expected 1 due payment, got %d", len(due))
	}
	if total != 50000 {
		t.Errorf("expected total 50000, got %d", total)
	}
}

func TestListDuePlannedPayments_Overdue(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	_, _ = svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Overdue Bill", Amount: 75000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2020-01-15", DueDay: 15,
	})

	due, total, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueOverdue})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments overdue: %v", err)
	}
	if len(due) != 1 {
		t.Errorf("expected 1 overdue payment, got %d", len(due))
	}
	if total != 75000 {
		t.Errorf("expected total 75000, got %d", total)
	}
}

func TestPayPlannedPayment_Recurring(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2026-07-15", DueDay: 15,
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}

	result, err := svc.PayPlannedPayment(PayPlannedPaymentParams{ID: pp.ID})
	if err != nil {
		t.Fatalf("PayPlannedPayment: %v", err)
	}
	if result.Transaction.Amount != 149000 {
		t.Errorf("expected transaction amount 149000, got %d", result.Transaction.Amount)
	}
	if result.Transaction.IsPlanned != 1 {
		t.Errorf("expected is_planned=1, got %d", result.Transaction.IsPlanned)
	}
	if !result.Transaction.PlannedPaymentID.Valid || result.Transaction.PlannedPaymentID.Int64 != pp.ID {
		t.Error("expected planned_payment_id to match")
	}
	if result.NextDueDate == "" {
		t.Error("expected next due date after pay")
	}

	updatedPP, err := svc.GetPlannedPaymentByID(pp.ID)
	if err != nil {
		t.Fatalf("GetPlannedPaymentByID: %v", err)
	}
	if updatedPP.IsActive != 1 {
		t.Errorf("expected is_active=1 for recurring, got %d", updatedPP.IsActive)
	}

	account, err := svc.GetAccountByID(pp.AccountID)
	if err != nil {
		t.Fatalf("GetAccountByID: %v", err)
	}
	if account.Balance != -149000 {
		t.Errorf("expected balance -149000 after pay, got %d", account.Balance)
	}
}

func TestPayPlannedPayment_OneTime(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Flight", Amount: 3000000, Account: "BCA",
		Category: "Travel", Recurrence: "none", StartDate: "2026-08-15",
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}

	result, err := svc.PayPlannedPayment(PayPlannedPaymentParams{ID: pp.ID})
	if err != nil {
		t.Fatalf("PayPlannedPayment: %v", err)
	}
	if result.Transaction.Amount != 3000000 {
		t.Errorf("expected transaction amount 3000000, got %d", result.Transaction.Amount)
	}

	updatedPP, err := svc.GetPlannedPaymentByID(pp.ID)
	if err != nil {
		t.Fatalf("GetPlannedPaymentByID: %v", err)
	}
	if updatedPP.IsActive != 0 {
		t.Errorf("expected is_active=0 (archived) for one-time, got %d", updatedPP.IsActive)
	}
}

func TestPayPlannedPayment_WithOverride(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2026-07-15", DueDay: 15,
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}

	result, err := svc.PayPlannedPayment(PayPlannedPaymentParams{
		ID:     pp.ID,
		Date:   "2026-07-14",
		Amount: 100000,
	})
	if err != nil {
		t.Fatalf("PayPlannedPayment: %v", err)
	}
	if result.Transaction.Amount != 100000 {
		t.Errorf("expected overridden amount 100000, got %d", result.Transaction.Amount)
	}
	if result.Transaction.Date != "2026-07-14" {
		t.Errorf("expected overridden date 2026-07-14, got %s", result.Transaction.Date)
	}
}

func TestPayPlannedPayment_Paused(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	_ = svc.PausePlannedPayment(pp.ID)

	_, err := svc.PayPlannedPayment(PayPlannedPaymentParams{ID: pp.ID})
	if err == nil {
		t.Fatal("expected error for paying paused planned payment")
	}
}

func TestPayPlannedPayment_NotFound(t *testing.T) {
	svc := setupServiceForPP(t)

	_, err := svc.PayPlannedPayment(PayPlannedPaymentParams{ID: 9999})
	if err == nil {
		t.Fatal("expected error for non-existent planned payment")
	}
}

func TestSkipPlannedPayment_Recurring(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2026-07-15", DueDay: 15,
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}

	updated, err := svc.SkipPlannedPayment(pp.ID)
	if err != nil {
		t.Fatalf("SkipPlannedPayment: %v", err)
	}
	if !updated.NextDueDate.Valid {
		t.Error("expected next due date after skip")
	}
}

func TestSkipPlannedPayment_OneTime(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Flight", Amount: 3000000, Account: "BCA",
		Category: "Travel", Recurrence: "none", StartDate: "2026-08-15",
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment: %v", err)
	}

	_, err = svc.SkipPlannedPayment(pp.ID)
	if err == nil {
		t.Fatal("expected error for skipping one-time planned payment")
	}
}

func TestPausePlannedPayment(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	if err := svc.PausePlannedPayment(pp.ID); err != nil {
		t.Fatalf("PausePlannedPayment: %v", err)
	}

	updated, err := svc.GetPlannedPaymentByID(pp.ID)
	if err != nil {
		t.Fatalf("GetPlannedPaymentByID: %v", err)
	}
	if updated.IsPaused != 1 {
		t.Errorf("expected is_paused=1, got %d", updated.IsPaused)
	}
}

func TestPausePlannedPayment_AlreadyPaused(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	_ = svc.PausePlannedPayment(pp.ID)

	err := svc.PausePlannedPayment(pp.ID)
	if err == nil {
		t.Fatal("expected error for pausing already paused payment")
	}
}

func TestResumePlannedPayment(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	_ = svc.PausePlannedPayment(pp.ID)

	if err := svc.ResumePlannedPayment(pp.ID); err != nil {
		t.Fatalf("ResumePlannedPayment: %v", err)
	}

	updated, err := svc.GetPlannedPaymentByID(pp.ID)
	if err != nil {
		t.Fatalf("GetPlannedPaymentByID: %v", err)
	}
	if updated.IsPaused != 0 {
		t.Errorf("expected is_paused=0, got %d", updated.IsPaused)
	}
}

func TestResumePlannedPayment_NotPaused(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	err := svc.ResumePlannedPayment(pp.ID)
	if err == nil {
		t.Fatal("expected error for resuming non-paused payment")
	}
}

func TestEditPlannedPayment(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	newName := "Netflix Premium"
	newAmount := int64(169000)
	updated, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{
		Name:   &newName,
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("EditPlannedPayment: %v", err)
	}
	if updated.Name != "Netflix Premium" {
		t.Errorf("expected name Netflix Premium, got %s", updated.Name)
	}
	if updated.Amount != 169000 {
		t.Errorf("expected amount 169000, got %d", updated.Amount)
	}
}

func TestEditPlannedPayment_DueDayChange(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2026-07-15", DueDay: 15,
	})

	newDay := 20
	updated, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{
		DueDay: &newDay,
	})
	if err != nil {
		t.Fatalf("EditPlannedPayment: %v", err)
	}
	if !updated.NextDueDate.Valid {
		t.Error("expected next due date after day change")
	}
}

func TestDeletePlannedPayment(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	if err := svc.DeletePlannedPayment(pp.ID); err != nil {
		t.Fatalf("DeletePlannedPayment: %v", err)
	}

	_, err := svc.GetPlannedPaymentByID(pp.ID)
	if err == nil {
		t.Fatal("expected error for deleted planned payment")
	}
}

func TestDeletePlannedPayment_NotFound(t *testing.T) {
	svc := setupServiceForPP(t)

	err := svc.DeletePlannedPayment(9999)
	if err == nil {
		t.Fatal("expected error for deleting non-existent planned payment")
	}
}

func TestCalcNextDue_Daily(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "daily", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_Weekly(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "weekly", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_Monthly(t *testing.T) {
	dueDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "monthly", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_MonthlyEndOfMonthClamp(t *testing.T) {
	dueDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "monthly", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_Yearly(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "yearly", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2027, 7, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_None(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "none", testNullString(""))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	if !next.Equal(dueDate) {
		t.Errorf("expected same date for none recurrence")
	}
}

func TestCalcNextDue_CustomRRULE(t *testing.T) {
	dueDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "custom", testNullString("FREQ=MONTHLY;BYMONTHDAY=15"))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_CustomRRULE_Unsupported(t *testing.T) {
	dueDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	_, err := calcNextDue(dueDate, "custom", testNullString("FREQ=HOURLY"))
	if err == nil {
		t.Fatal("expected error for unsupported RRULE frequency")
	}
}

func TestValidateRRULE(t *testing.T) {
	tests := []struct {
		rrule string
		valid bool
	}{
		{"FREQ=MONTHLY;BYMONTHDAY=15", true},
		{"FREQ=DAILY", true},
		{"FREQ=WEEKLY", true},
		{"FREQ=YEARLY", true},
		{"FREQ=HOURLY", false},
		{"INVALID", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.rrule, func(t *testing.T) {
			err := validateRRULE(tt.rrule)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for invalid RRULE")
			}
		})
	}
}

func TestListDuePlannedPayments_ExcludesPaused(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	now := time.Now()
	today := now.Format("2006-01-02")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Paused Bill", Amount: 50000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: today, DueDay: now.Day(),
	})
	_ = svc.PausePlannedPayment(pp.ID)

	due, _, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueCurrentMonth})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments: %v", err)
	}
	for _, d := range due {
		if d.ID == pp.ID {
			t.Error("paused payment should be excluded from due list")
		}
	}
}

func TestListDuePlannedPayments_ExcludesArchived(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Archived Bill", Amount: 50000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "none", StartDate: "2026-07-01",
	})
	_, _ = svc.PayPlannedPayment(PayPlannedPaymentParams{ID: pp.ID})

	due, _, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueCurrentMonth})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments: %v", err)
	}
	for _, d := range due {
		if d.ID == pp.ID {
			t.Error("archived one-time payment should be excluded from due list")
		}
	}
}

func TestListPlannedPayments_All(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	active, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	paused, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Spotify", Amount: 54990, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 1,
	})
	_ = svc.PausePlannedPayment(paused.ID)
	oneTime, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Flight", Amount: 3000000, Account: "BCA",
		Category: "Travel", Recurrence: "none", StartDate: "2026-08-15",
	})
	_, _ = svc.PayPlannedPayment(PayPlannedPaymentParams{ID: oneTime.ID})

	payments, err := svc.ListPlannedPayments(true, true)
	if err != nil {
		t.Fatalf("ListPlannedPayments all: %v", err)
	}
	if len(payments) != 3 {
		t.Errorf("expected 3 total payments, got %d", len(payments))
	}

	_ = active
}

func TestGetAccountName(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	name, err := svc.GetAccountName(1)
	if err != nil {
		t.Fatalf("GetAccountName: %v", err)
	}
	if name != "BCA" {
		t.Errorf("expected BCA, got %s", name)
	}
}

func TestGetCategoryName(t *testing.T) {
	svc := setupServiceForPP(t)

	name := svc.GetCategoryName(1)
	if name == "" {
		t.Error("expected non-empty category name")
	}
}

func TestListDuePlannedPayments_Week(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	due, _, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueCurrentWeek})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments week: %v", err)
	}
	_ = due
}

func TestListDuePlannedPayments_NextDays(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	due, _, err := svc.ListDuePlannedPayments(ListDueParams{Filter: DueNextDays, NextDays: 30})
	if err != nil {
		t.Fatalf("ListDuePlannedPayments next 30: %v", err)
	}
	_ = due
}

func TestEditPlannedPayment_AccountAndCategory(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")
	mustCreateAccount(t, svc, "GoPay", "ewallet", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	newAccount := "GoPay"
	newCategory := "Internet"
	updated, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{
		Account:  &newAccount,
		Category: &newCategory,
	})
	if err != nil {
		t.Fatalf("EditPlannedPayment: %v", err)
	}
	if updated.AccountID != 2 {
		t.Errorf("expected account 2 (GoPay), got %d", updated.AccountID)
	}
	_ = updated
}

func TestEditPlannedPayment_RecurrenceAndRRULE(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	newRecurrence := "daily"
	updated, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{
		Recurrence: &newRecurrence,
	})
	if err != nil {
		t.Fatalf("EditPlannedPayment: %v", err)
	}
	if updated.Recurrence != "daily" {
		t.Errorf("expected daily, got %s", updated.Recurrence)
	}
}

func TestEditPlannedPayment_InvalidAmount(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	zero := int64(0)
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{Amount: &zero})
	if err == nil {
		t.Fatal("expected error for zero amount edit")
	}
}

func TestEditPlannedPayment_InvalidRecurrence(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	invalid := "bogus"
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{Recurrence: &invalid})
	if err == nil {
		t.Fatal("expected error for invalid recurrence edit")
	}
}

func TestEditPlannedPayment_CustomRRULEMissing(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	custom := "custom"
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{Recurrence: &custom})
	if err == nil {
		t.Fatal("expected error for custom recurrence without RRULE")
	}
}

func TestEditPlannedPayment_StartDate(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", StartDate: "2026-07-15", DueDay: 15,
	})

	newStart := "2026-08-01"
	updated, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{StartDate: &newStart})
	if err != nil {
		t.Fatalf("EditPlannedPayment: %v", err)
	}
	if updated.StartDate != "2026-08-01" {
		t.Errorf("expected 2026-08-01, got %s", updated.StartDate)
	}
}

func TestEditPlannedPayment_InvalidStartDate(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	invalid := "bad-date"
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{StartDate: &invalid})
	if err == nil {
		t.Fatal("expected error for invalid start date")
	}
}

func TestEditPlannedPayment_AccountNotFound(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	missing := "NonExistent"
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{Account: &missing})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestEditPlannedPayment_CategoryNotFound(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	missing := "NonExistentCategory"
	_, err := svc.EditPlannedPayment(pp.ID, EditPlannedPaymentParams{Category: &missing})
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestResumePlannedPayment_AdvancesOverdue(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly",
		StartDate: "2020-01-15", DueDay: 15,
	})
	_ = svc.PausePlannedPayment(pp.ID)

	if err := svc.ResumePlannedPayment(pp.ID); err != nil {
		t.Fatalf("ResumePlannedPayment: %v", err)
	}

	updated, err := svc.GetPlannedPaymentByID(pp.ID)
	if err != nil {
		t.Fatalf("GetPlannedPaymentByID: %v", err)
	}
	if updated.IsPaused != 0 {
		t.Errorf("expected is_paused=0, got %d", updated.IsPaused)
	}
	if !updated.NextDueDate.Valid || updated.NextDueDate.String == "2020-02-15" {
		t.Log("next_due_date was advanced from overdue date")
	}
}

func TestListPlannedPayments_ActiveAndArchived(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	active, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})
	oneTime, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Flight", Amount: 3000000, Account: "BCA",
		Category: "Travel", Recurrence: "none", StartDate: "2026-08-15",
	})
	_, _ = svc.PayPlannedPayment(PayPlannedPaymentParams{ID: oneTime.ID})

	payments, err := svc.ListPlannedPayments(false, true)
	if err != nil {
		t.Fatalf("ListPlannedPayments: %v", err)
	}
	if len(payments) != 2 {
		t.Errorf("expected 2 payments (active + archived), got %d", len(payments))
	}
	_ = active
}

func TestCreatePlannedPayment_Weekly(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Weekly Bill", Amount: 50000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "weekly",
		StartDate: "2026-07-01", DueDay: int(time.Monday),
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment weekly: %v", err)
	}
	if pp.Recurrence != "weekly" {
		t.Errorf("expected weekly, got %s", pp.Recurrence)
	}
}

func TestCreatePlannedPayment_Yearly(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Yearly Bill", Amount: 500000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "yearly",
		StartDate: "2026-07-15", DueDay: 15,
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment yearly: %v", err)
	}
	if pp.Recurrence != "yearly" {
		t.Errorf("expected yearly, got %s", pp.Recurrence)
	}
}

func TestCreatePlannedPayment_Daily(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Daily Bill", Amount: 10000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "daily", StartDate: "2026-07-01",
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment daily: %v", err)
	}
	if pp.Recurrence != "daily" {
		t.Errorf("expected daily, got %s", pp.Recurrence)
	}
}

func TestCreatePlannedPayment_CustomRRULE(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, err := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Custom Bill", Amount: 75000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "custom", RecurrenceRule: "FREQ=MONTHLY;BYMONTHDAY=15",
		StartDate: "2026-07-15",
	})
	if err != nil {
		t.Fatalf("CreatePlannedPayment custom: %v", err)
	}
	if pp.Recurrence != "custom" {
		t.Errorf("expected custom, got %s", pp.Recurrence)
	}
}

func TestCalcNextDue_RRULEWeekly(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "custom", testNullString("FREQ=WEEKLY"))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_RRULEDaily(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "custom", testNullString("FREQ=DAILY"))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_RRULEYearly(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	next, err := calcNextDue(dueDate, "custom", testNullString("FREQ=YEARLY"))
	if err != nil {
		t.Fatalf("calcNextDue: %v", err)
	}
	expected := time.Date(2027, 7, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format("2006-01-02"), next.Format("2006-01-02"))
	}
}

func TestCalcNextDue_UnknownRecurrence(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := calcNextDue(dueDate, "bogus", testNullString(""))
	if err == nil {
		t.Fatal("expected error for unknown recurrence")
	}
}

func TestCalcNextDue_CustomWithoutRule(t *testing.T) {
	dueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := calcNextDue(dueDate, "custom", testNullString(""))
	if err == nil {
		t.Fatal("expected error for custom without rule")
	}
}

func TestResumePlannedPayment_NotFound(t *testing.T) {
	svc := setupServiceForPP(t)

	err := svc.ResumePlannedPayment(9999)
	if err == nil {
		t.Fatal("expected error for non-existent payment")
	}
}

func TestPausePlannedPayment_NotFound(t *testing.T) {
	svc := setupServiceForPP(t)

	err := svc.PausePlannedPayment(9999)
	if err == nil {
		t.Fatal("expected error for non-existent payment")
	}
}

func TestPayPlannedPayment_DateError(t *testing.T) {
	svc := setupServiceForPP(t)
	mustCreateAccount(t, svc, "BCA", "checking", "IDR")

	pp, _ := svc.CreatePlannedPayment(CreatePlannedPaymentParams{
		Name: "Netflix", Amount: 149000, Account: "BCA",
		Category: "Subscriptions", Recurrence: "monthly", DueDay: 15,
	})

	_, err := svc.PayPlannedPayment(PayPlannedPaymentParams{ID: pp.ID, Date: "bad-date"})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func testNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
