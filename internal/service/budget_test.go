package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/testdb"
)

func TestSetBudgetMonthlyWithCategories(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Monthly Food",
		Amount:     2000000,
		Period:     "monthly",
		Categories: []string{"Restaurant", "Groceries"},
		NotifyPct:  80,
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}
	if result.Budget.Amount != 2000000 {
		t.Errorf("expected amount 2000000, got %d", result.Budget.Amount)
	}
	if result.Budget.Type != "monthly" {
		t.Errorf("expected type monthly, got %s", result.Budget.Type)
	}
	if len(result.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result.Categories))
	}
}

func TestSetBudgetWithTags(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateTag("japan-2026")
	_, _ = svc.CreateTag("tokyo")

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Japan Trip",
		Amount:     10000000,
		Period:     "one_time",
		From:       "2026-01-01",
		To:         "2026-12-31",
		Tags:       []string{"japan-2026", "tokyo"},
	})
	if err != nil {
		t.Fatalf("SetBudget with tags: %v", err)
	}
	if len(result.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result.Tags))
	}
}

func TestSetBudgetMixedTargets(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateTag("travel")

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Travel Budget",
		Amount:     5000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
		Tags:       []string{"travel"},
	})
	if err != nil {
		t.Fatalf("SetBudget mixed: %v", err)
	}
	if len(result.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(result.Categories))
	}
	if len(result.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(result.Tags))
	}
}

func TestSetBudgetUpsertSameNameAndPeriod(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Monthly Food",
		Amount:     2000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("first SetBudget: %v", err)
	}

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Monthly Food",
		Amount:     2500000,
		Period:     "monthly",
		Categories: []string{"Coffee & Snacks"},
		NotifyPct:  75,
	})
	if err != nil {
		t.Fatalf("second SetBudget (upsert): %v", err)
	}
	if result.Budget.Amount != 2500000 {
		t.Errorf("expected updated amount 2500000, got %d", result.Budget.Amount)
	}
	if len(result.Categories) != 1 {
		t.Errorf("expected 1 category after upsert, got %d", len(result.Categories))
	}
	if result.Categories[0].Name != "Coffee & Snacks" {
		t.Errorf("expected category Coffee & Snacks, got %s", result.Categories[0].Name)
	}
}

func TestSetBudgetNoTargets(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:   "Untargeted",
		Amount: 1000000,
		Period: "monthly",
	})
	if err == nil {
		t.Fatal("expected error for budget without targets")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestSetBudgetInvalidAmount(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Zero Budget",
		Amount:     0,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = svc.SetBudget(SetBudgetParams{
		Name:       "Negative Budget",
		Amount:     -100,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestSetBudgetInvalidPeriod(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Daily Budget",
		Amount:     1000000,
		Period:     "daily",
		Categories: []string{"Restaurant"},
	})
	if err == nil {
		t.Fatal("expected error for unsupported period")
	}
	if !strings.Contains(err.Error(), "supported periods") {
		t.Errorf("expected 'supported periods' in error, got %v", err)
	}
}

func TestSetBudgetOneTimeWithoutDates(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Trip",
		Amount:     10000000,
		Period:     "one_time",
		Categories: []string{"Restaurant"},
	})
	if err == nil {
		t.Fatal("expected error for one_time without dates")
	}
	if !strings.Contains(err.Error(), "requires --from and --to") {
		t.Errorf("expected 'requires --from and --to' in error, got %v", err)
	}
}

func TestSetBudgetInvalidNotifyPct(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Bad Notify",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
		NotifyPct:  101,
	})
	if err == nil {
		t.Fatal("expected error for notify > 100")
	}
}

func TestSetBudgetCategoryNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Missing Cat",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"GhostCategory"},
	})
	if err == nil {
		t.Fatal("expected error for missing category")
	}
	if !strings.Contains(err.Error(), "GhostCategory") {
		t.Errorf("expected 'GhostCategory' in error, got %v", err)
	}
}

func TestSetBudgetTagNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:   "Missing Tag",
		Amount: 1000000,
		Period: "monthly",
		Tags:   []string{"GhostTag"},
	})
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
	if !strings.Contains(err.Error(), "GhostTag") {
		t.Errorf("expected 'GhostTag' in error, got %v", err)
	}
}

func TestSetBudgetExplicitDates(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Custom Period",
		Amount:     500000,
		Period:     "monthly",
		From:       "2026-07-15",
		To:         "2026-08-14",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget with explicit dates: %v", err)
	}
	if result.Budget.PeriodStart != "2026-07-15" {
		t.Errorf("expected period_start 2026-07-15, got %s", result.Budget.PeriodStart)
	}
	if result.Budget.PeriodEnd != "2026-08-14" {
		t.Errorf("expected period_end 2026-08-14, got %s", result.Budget.PeriodEnd)
	}
}

func TestPeriodCalculationDefault(t *testing.T) {
	start, end, err := calculatePeriod("monthly", "", "")
	if err != nil {
		t.Fatalf("calculatePeriod monthly: %v", err)
	}
	if start == "" || end == "" {
		t.Error("expected non-empty start and end for monthly")
	}

	start, end, err = calculatePeriod("weekly", "", "")
	if err != nil {
		t.Fatalf("calculatePeriod weekly: %v", err)
	}
	if start == "" || end == "" {
		t.Error("expected non-empty start and end for weekly")
	}

	start, end, err = calculatePeriod("yearly", "", "")
	if err != nil {
		t.Fatalf("calculatePeriod yearly: %v", err)
	}
	if start == "" || end == "" {
		t.Error("expected non-empty start and end for yearly")
	}
}

func TestPeriodCalculationExplicit(t *testing.T) {
	start, end, err := calculatePeriod("monthly", "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("calculatePeriod explicit: %v", err)
	}
	if start != "2026-01-01" {
		t.Errorf("expected start 2026-01-01, got %s", start)
	}
	if end != "2026-01-31" {
		t.Errorf("expected end 2026-01-31, got %s", end)
	}
}

func TestPeriodCalculationInvalidDate(t *testing.T) {
	_, _, err := calculatePeriod("monthly", "invalid", "2026-01-31")
	if err == nil {
		t.Fatal("expected error for invalid from date")
	}

	_, _, err = calculatePeriod("monthly", "2026-01-01", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid to date")
	}
}

func TestListBudgetsActive(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Active Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	items, err := svc.ListBudgets(ListBudgetsParams{All: false})
	if err != nil {
		t.Fatalf("ListBudgets: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 active budget, got %d", len(items))
	}
}

func TestListBudgetsWithSpent(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	budget, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      200000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	items, err := svc.ListBudgets(ListBudgetsParams{All: false})
	if err != nil {
		t.Fatalf("ListBudgets: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 budget, got %d", len(items))
	}
	if items[0].Spent != 200000 {
		t.Errorf("expected spent 200000, got %d", items[0].Spent)
	}
	expectedRemaining := budget.Budget.Amount - int64(200000)
	if items[0].Remaining != expectedRemaining {
		t.Errorf("expected remaining %d, got %d", expectedRemaining, items[0].Remaining)
	}
}

func TestListBudgetsInactiveExcluded(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	if err := svc.RemoveBudget(result.Budget.ID); err != nil {
		t.Fatalf("RemoveBudget: %v", err)
	}

	items, err := svc.ListBudgets(ListBudgetsParams{All: false})
	if err != nil {
		t.Fatalf("ListBudgets active: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 active budgets, got %d", len(items))
	}

	items, err = svc.ListBudgets(ListBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("ListBudgets all: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 budget with --all, got %d", len(items))
	}
}

func TestCheckBudgetsAll(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 check result")
	}
	r := results[0]
	if r.Status != BudgetStatusOK {
		t.Errorf("expected status ok, got %s", r.Status)
	}
	if r.Spent != 0 {
		t.Errorf("expected spent 0, got %d", r.Spent)
	}
}

func TestCheckBudgetsSingle(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{Identifier: "Food Budget"})
	if err != nil {
		t.Fatalf("CheckBudgets by name: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Budget.ID != result.Budget.ID {
		t.Errorf("expected budget ID %d, got %d", result.Budget.ID, results[0].Budget.ID)
	}

	idStr := fmt.Sprintf("%d", result.Budget.ID)
	results, err = svc.CheckBudgets(CheckBudgetsParams{Identifier: idStr})
	if err != nil {
		t.Fatalf("CheckBudgets by ID: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result by ID, got %d", len(results))
	}
}

func TestCheckBudgetsNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CheckBudgets(CheckBudgetsParams{Identifier: "Ghost Budget"})
	if err == nil {
		t.Fatal("expected error for missing budget")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCheckBudgetsStatusWarning(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
		NotifyPct:  80,
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      800000,
		Description: "Expensive dinner",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Status != BudgetStatusWarning {
		t.Errorf("expected status warning, got %s (spent: %d, limit: %d)",
			results[0].Status, results[0].Spent, result.Budget.Amount)
	}
}

func TestCheckBudgetsStatusOver(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      1000000,
		Description: "Maxed out",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Status != BudgetStatusOver {
		t.Errorf("expected status over, got %s", results[0].Status)
	}
}

func TestSpendingExcludesNonExpense(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Salary"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	_, err = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Gaji",
		Category:    "Salary",
		Account:     "BCA",
	})
	if err != nil {
		t.Fatalf("AddIncome: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Spent != 0 {
		t.Errorf("expected spent 0 (income excluded), got %d", results[0].Spent)
	}
}

func TestSpendingExcludesArchived(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food Budget",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	txn, err := svc.AddExpense(CreateExpenseParams{
		Amount:      200000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	if err := svc.RemoveTransaction(txn.Transaction.ID); err != nil {
		t.Fatalf("RemoveTransaction: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Spent != 0 {
		t.Errorf("expected spent 0 (archived excluded), got %d", results[0].Spent)
	}
}

func TestSpendingMixedOverlapDoubleCounted(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("japan-2026")

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Japan Budget",
		Amount:     10000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
		Tags:       []string{"japan-2026"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      500000,
		Description: "Sushi",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"japan-2026"},
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Spent != 1000000 {
		t.Errorf("expected spent 1000000 (double-counted 500000), got %d", results[0].Spent)
	}
}

func TestRecurringAutoGeneration(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Monthly Bills",
		Amount:     500000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
		NotifyPct:  75,
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected check results")
	}

	allBudgets, err := svc.ListBudgets(ListBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("ListBudgets all: %v", err)
	}
	if len(allBudgets) != 1 {
		t.Errorf("expected 1 budget (no auto-gen because no prior period exists), got %d", len(allBudgets))
	}
}

func TestRecurringOneTimeExcluded(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "One Time Trip",
		Amount:     1000000,
		Period:     "one_time",
		From:       "2025-01-01",
		To:         "2025-12-31",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget one_time: %v", err)
	}

	results, err := svc.CheckBudgets(CheckBudgetsParams{All: true})
	if err != nil {
		t.Fatalf("CheckBudgets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected one_time budget in check results")
	}
	if results[0].Budget.Type != "one_time" {
		t.Errorf("expected one_time budget type, got %s", results[0].Budget.Type)
	}
}

func TestEditBudgetAmount(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	newAmount := int64(2500000)
	edited, err := svc.EditBudget(result.Budget.ID, EditBudgetParams{
		Amount:    &newAmount,
		NotifyPct: int64Ptr(75),
	})
	if err != nil {
		t.Fatalf("EditBudget: %v", err)
	}
	if edited.Budget.Amount != 2500000 {
		t.Errorf("expected amount 2500000, got %d", edited.Budget.Amount)
	}
}

func TestEditBudgetName(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	edited, err := svc.EditBudget(result.Budget.ID, EditBudgetParams{
		Name: "Monthly Essentials",
	})
	if err != nil {
		t.Fatalf("EditBudget: %v", err)
	}
	if !edited.Budget.Name.Valid || edited.Budget.Name.String != "Monthly Essentials" {
		t.Errorf("expected name 'Monthly Essentials', got %v", edited.Budget.Name)
	}
}

func TestEditBudgetTargets(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateTag("food-tag")

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	edited, err := svc.EditBudget(result.Budget.ID, EditBudgetParams{
		AddCategories:    []string{"Coffee & Snacks"},
		RemoveCategories: []string{"Restaurant"},
		AddTags:          []string{"food-tag"},
	})
	if err != nil {
		t.Fatalf("EditBudget: %v", err)
	}
	if len(edited.Categories) != 1 || edited.Categories[0].Name != "Coffee & Snacks" {
		t.Errorf("expected 1 category Coffee & Snacks, got %v", edited.Categories)
	}
	if len(edited.Tags) != 1 || edited.Tags[0].Name != "food-tag" {
		t.Errorf("expected 1 tag food-tag, got %v", edited.Tags)
	}
}

func TestEditBudgetNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.EditBudget(99, EditBudgetParams{
		Name: "Ghost",
	})
	if err == nil {
		t.Fatal("expected error for missing budget")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestRemoveBudget(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	if err := svc.RemoveBudget(result.Budget.ID); err != nil {
		t.Fatalf("RemoveBudget: %v", err)
	}

	items, err := svc.ListBudgets(ListBudgetsParams{All: false})
	if err != nil {
		t.Fatalf("ListBudgets: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 active budgets after removal, got %d", len(items))
	}
}

func TestRemoveBudgetNotFound(t *testing.T) {
	svc := setupService(t)

	err := svc.RemoveBudget(99)
	if err == nil {
		t.Fatal("expected error for missing budget")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestEditBudgetInvalidAmount(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	negAmount := int64(-1)
	_, err = svc.EditBudget(result.Budget.ID, EditBudgetParams{Amount: &negAmount})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	zeroAmount := int64(0)
	_, err = svc.EditBudget(result.Budget.ID, EditBudgetParams{Amount: &zeroAmount})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for zero, got %v", err)
	}
}

func TestSetBudgetWeekly(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Weekly Groceries",
		Amount:     500000,
		Period:     "weekly",
		Categories: []string{"Groceries"},
	})
	if err != nil {
		t.Fatalf("SetBudget weekly: %v", err)
	}
	if result.Budget.Type != "weekly" {
		t.Errorf("expected type weekly, got %s", result.Budget.Type)
	}
}

func TestSetBudgetYearly(t *testing.T) {
	svc := setupService(t)

	result, err := svc.SetBudget(SetBudgetParams{
		Name:       "Yearly Travel",
		Amount:     50000000,
		Period:     "yearly",
		Categories: []string{"Restaurant"},
	})
	if err != nil {
		t.Fatalf("SetBudget yearly: %v", err)
	}
	if result.Budget.Type != "yearly" {
		t.Errorf("expected type yearly, got %s", result.Budget.Type)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

// Mock queriers for error path coverage
type budgetCreateFailQuerier struct {
	gen.Querier
}

func (q *budgetCreateFailQuerier) CreateBudget(ctx context.Context, arg gen.CreateBudgetParams) (*gen.Budget, error) {
	return nil, fmt.Errorf("mock create failure")
}

type budgetUpdateFailQuerier struct {
	gen.Querier
}

func (q *budgetUpdateFailQuerier) UpdateBudget(ctx context.Context, arg gen.UpdateBudgetParams) (*gen.Budget, error) {
	return nil, fmt.Errorf("mock update failure")
}

type budgetGetFailQuerier struct {
	gen.Querier
}

func (q *budgetGetFailQuerier) GetBudgetByID(ctx context.Context, id int64) (*gen.Budget, error) {
	return nil, fmt.Errorf("mock get failure")
}

type budgetListFailQuerier struct {
	gen.Querier
}

func (q *budgetListFailQuerier) ListActiveBudgets(ctx context.Context) ([]*gen.Budget, error) {
	return nil, fmt.Errorf("mock list failure")
}

type budgetSpendingFailQuerier struct {
	gen.Querier
	catFail bool
}

func (q *budgetSpendingFailQuerier) SumCategoryExpenses(ctx context.Context, arg gen.SumCategoryExpensesParams) (interface{}, error) {
	if q.catFail {
		return nil, fmt.Errorf("mock spending failure")
	}
	return int64(0), nil
}

func (q *budgetSpendingFailQuerier) SumTagExpenses(ctx context.Context, arg gen.SumTagExpensesParams) (interface{}, error) {
	return nil, fmt.Errorf("mock tag spending failure")
}

func TestSetBudgetCreateError(t *testing.T) {
	dbase := testdb.Open(t)
	svc := NewWithQuerier(dbase, &budgetCreateFailQuerier{Querier: gen.New(dbase)})

	_, err := svc.SetBudget(SetBudgetParams{
		Name:       "Food",
		Amount:     1000000,
		Period:     "monthly",
		Categories: []string{"Food & Dining"},
	})
	if err == nil {
		t.Fatal("expected create error")
	}
	if !strings.Contains(err.Error(), "mock create failure") {
		t.Errorf("expected mock error, got %v", err)
	}
}

func TestEditBudgetGetError(t *testing.T) {
	dbase := testdb.Open(t)
	svc := NewWithQuerier(dbase, &budgetGetFailQuerier{Querier: gen.New(dbase)})

	_, err := svc.EditBudget(1, EditBudgetParams{Name: "Test"})
	if err == nil {
		t.Fatal("expected get error")
	}
	if !strings.Contains(err.Error(), "mock get failure") {
		t.Errorf("expected mock error, got %v", err)
	}
}

func TestListBudgetsSpendingError(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateBudget(context.Background(), gen.CreateBudgetParams{
		Name:        sql.NullString{String: "Test", Valid: true},
		Amount:      1000000,
		Currency:    "IDR",
		Type:        "monthly",
		PeriodStart: "2026-07-01",
		PeriodEnd:   "2026-07-31",
	})
	svc := NewWithQuerier(dbase, &budgetSpendingFailQuerier{Querier: q, catFail: true})

	_, err := svc.ListBudgets(ListBudgetsParams{All: false})
	if err == nil {
		t.Fatal("expected spending error")
	}
}
