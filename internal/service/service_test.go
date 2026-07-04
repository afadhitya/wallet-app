package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/testdb"
	"github.com/afadhitya/wallet-app/pkg/config"
)

func setupService(t *testing.T) *Service {
	t.Helper()
	return New(testdb.Open(t))
}

func TestAddExpense(t *testing.T) {
	svc := setupService(t)

	account, err := svc.CreateAccount("BCA", "checking", "IDR")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch at Warung",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	if result.Transaction.Amount != 35000 {
		t.Errorf("expected amount 35000, got %d", result.Transaction.Amount)
	}
	if result.Transaction.Type != "expense" {
		t.Errorf("expected type expense, got %s", result.Transaction.Type)
	}

	updated, err := svc.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -35000 {
		t.Errorf("expected balance -35000, got %d", updated.Balance)
	}
}

func TestAddExpenseWithTags(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("lunch")
	_, _ = svc.CreateTag("work")

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Team lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"lunch", "work"},
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	tags, err := svc.ListTransactionTags(result.Transaction.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestAddExpenseInvalidAmount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      0,
		Description: "Free",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      -100,
		Description: "Negative",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative amount, got %v", err)
	}
}

func TestAddExpenseMissingAccount(t *testing.T) {
	svc := setupService(t)

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for missing account")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestAddExpenseMissingCategory(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "GhostCategory",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestAddExpenseMissingTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestAddIncome(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")

	result, err := svc.AddIncome(CreateIncomeParams{
		Amount:      5000000,
		Description: "Gaji Juli",
		Category:    "Salary",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddIncome: %v", err)
	}

	if result.Transaction.Type != "income" {
		t.Errorf("expected type income, got %s", result.Transaction.Type)
	}

	updated, _ := svc.GetAccountByID(account.ID)
	if updated.Balance != 5000000 {
		t.Errorf("expected balance 5000000, got %d", updated.Balance)
	}
}

func TestAddTransfer(t *testing.T) {
	svc := setupService(t)

	src, _ := svc.CreateAccount("BCA", "checking", "IDR")
	dst, _ := svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddTransfer: %v", err)
	}

	if result.Transaction.Type != "transfer" {
		t.Errorf("expected type transfer, got %s", result.Transaction.Type)
	}
	if !result.Transaction.TransferToID.Valid || result.Transaction.TransferToID.Int64 != dst.ID {
		t.Errorf("expected transfer_to_id %d", dst.ID)
	}

	srcUpdated, _ := svc.GetAccountByID(src.ID)
	if srcUpdated.Balance != 800000 {
		t.Errorf("expected source balance 800000, got %d", srcUpdated.Balance)
	}

	dstUpdated, _ := svc.GetAccountByID(dst.ID)
	if dstUpdated.Balance != 200000 {
		t.Errorf("expected destination balance 200000, got %d", dstUpdated.Balance)
	}
}

func TestAddTransferInsufficientBalance(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	result, err := svc.AddTransfer(CreateTransferParams{
		Amount:      500000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err != nil {
		t.Fatalf("AddTransfer should succeed even with warning: %v", err)
	}

	if result.Warning == "" {
		t.Error("expected insufficient balance warning")
	}
}

func TestAddTransferSameAccount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "BCA",
	})
	if err == nil {
		t.Fatal("expected error for same source/destination")
	}
}

func TestEditTransaction(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	newAmount := int64(40000)
	result, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("EditTransaction: %v", err)
	}

	if result.Transaction.Amount != 40000 {
		t.Errorf("expected amount 40000, got %d", result.Transaction.Amount)
	}

	updated, _ := svc.GetAccountByID(txn.Transaction.AccountID)
	if updated.Balance != -40000 {
		t.Errorf("expected balance -40000, got %d", updated.Balance)
	}
}

func TestEditTransactionTags(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("lunch")
	_, _ = svc.CreateTag("work")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"lunch"},
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AddTagNames:    []string{"work"},
		RemoveTagNames: []string{"lunch"},
	})
	if err != nil {
		t.Fatalf("EditTransaction tags: %v", err)
	}

	tags, _ := svc.ListTransactionTags(txn.Transaction.ID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "work" {
		t.Errorf("expected tag 'work', got '%s'", tags[0].Name)
	}
}

func TestRemoveTransaction(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      5000000,
		Description: "Income",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      1000000,
		Description: "Big expense",
		Category:    "Restaurant",
		Account:     "BCA",
	})

	err := svc.RemoveTransaction(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("RemoveTransaction: %v", err)
	}

	updated, _ := svc.GetAccountByID(account.ID)
	if updated.Balance != 5000000 {
		t.Errorf("expected balance 5000000 after removal, got %d", updated.Balance)
	}
}

func TestRemoveTransactionNotFound(t *testing.T) {
	svc := setupService(t)

	err := svc.RemoveTransaction(9999)
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestAdjustBalance(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1500000,
		Description: "Correction",
	})
	if err != nil {
		t.Fatalf("AdjustBalance: %v", err)
	}

	if result.Difference != 500000 {
		t.Errorf("expected difference 500000, got %d", result.Difference)
	}

	updated, _ := svc.GetAccountByID(account.ID)
	if updated.Balance != 1500000 {
		t.Errorf("expected balance 1500000, got %d", updated.Balance)
	}
}

func TestBalanceRecalculationIgnoreArchived(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")

	txn1, _ := svc.AddIncome(CreateIncomeParams{
		Amount:      5000000,
		Description: "Income",
		Category:    "Salary",
		Account:     "BCA",
	})
	txn2, _ := svc.AddIncome(CreateIncomeParams{
		Amount:      3000000,
		Description: "Bonus",
		Category:    "Salary",
		Account:     "BCA",
	})

	_ = svc.RemoveTransaction(txn1.Transaction.ID)

	updated, _ := svc.GetAccountByID(account.ID)
	if updated.Balance != 3000000 {
		t.Errorf("expected balance 3000000 after archiving income, got %d", updated.Balance)
	}
	_ = txn2
}

func TestEditTransactionNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.EditTransaction(9999, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
}

func TestAdjustBalanceNoChange(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	result, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      0,
		Description: "No change",
	})
	if err != nil {
		t.Fatalf("AdjustBalance no change: %v", err)
	}
	if result.Difference != 0 {
		t.Errorf("expected difference 0, got %d", result.Difference)
	}
}

func TestCreateCategory(t *testing.T) {
	svc := setupService(t)

	cat, err := svc.CreateCategory("Hobby", "", "🎨")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if cat.Name != "Hobby" {
		t.Errorf("expected name 'Hobby', got '%s'", cat.Name)
	}
}

func TestCreateCategoryDuplicate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateCategory("Hobby", "", "")
	_, err := svc.CreateCategory("Hobby", "", "")
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestCreateCategoryWithParent(t *testing.T) {
	svc := setupService(t)

	parent, _ := svc.CreateCategory("Parent", "", "")
	parentIDStr := fmt.Sprintf("%d", parent.ID)
	child, err := svc.CreateCategory("Child", parentIDStr, "")
	if err != nil {
		t.Fatalf("CreateCategory with parent: %v", err)
	}
	if !child.ParentID.Valid || child.ParentID.Int64 != parent.ID {
		t.Errorf("expected parent_id %d, got %v", parent.ID, child.ParentID)
	}
}

func TestCreateCategoryMissingParent(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CreateCategory("Orphan", "999", "")
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestUpdateCategory(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("Old", "", "📦")
	updated, err := svc.UpdateCategory(cat.ID, "New", "📦📦")
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if updated.Name != "New" {
		t.Errorf("expected name 'New', got '%s'", updated.Name)
	}
}

func TestUpdateCategoryNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.UpdateCategory(999, "Ghost", "")
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestArchiveCategory(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("Temp", "", "")
	err := svc.ArchiveCategory(cat.ID)
	if err != nil {
		t.Fatalf("ArchiveCategory: %v", err)
	}

	active, _ := svc.ListCategories()
	for _, c := range active {
		if c.ID == cat.ID {
			t.Error("expected archived category to not appear in ListCategories")
		}
	}
}

func TestArchiveCategoryNotFound(t *testing.T) {
	svc := setupService(t)

	err := svc.ArchiveCategory(999)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestListCategoriesActive(t *testing.T) {
	svc := setupService(t)

	categories, err := svc.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	archivedCount := 0
	for _, c := range categories {
		if c.IsArchived != 0 {
			archivedCount++
		}
	}
	if archivedCount > 0 {
		t.Errorf("expected only active categories, got %d archived", archivedCount)
	}
}

func TestCreateTag(t *testing.T) {
	svc := setupService(t)

	tag, err := svc.CreateTag("japan-2026")
	if err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if tag.Name != "japan-2026" {
		t.Errorf("expected name 'japan-2026', got '%s'", tag.Name)
	}
}

func TestCreateTagDuplicate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateTag("unique")
	_, err := svc.CreateTag("unique")
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestCreateTagEmptyName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CreateTag("")
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestDeleteTag(t *testing.T) {
	svc := setupService(t)

	tag, _ := svc.CreateTag("temp")
	err := svc.DeleteTag(tag.ID)
	if err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}

	tags, _ := svc.ListTags()
	for _, tagItem := range tags {
		if tagItem.ID == tag.ID {
			t.Error("expected deleted tag to not appear in ListTags")
		}
	}
}

func TestDeleteTagNotFound(t *testing.T) {
	svc := setupService(t)

	err := svc.DeleteTag(999)
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestListTags(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateTag("alpha")
	_, _ = svc.CreateTag("beta")

	tags, err := svc.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestListTransactionTags(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("food")
	_, _ = svc.CreateTag("dinner")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Dinner",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"food", "dinner"},
	})

	tags, err := svc.ListTransactionTags(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("ListTransactionTags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestResolveAccountByID(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")
	resolved, err := svc.ResolveAccount("1")
	if err != nil {
		t.Fatalf("ResolveAccount by ID: %v", err)
	}
	if resolved.ID != account.ID {
		t.Errorf("expected ID %d, got %d", account.ID, resolved.ID)
	}
}

func TestResolveAccountByName(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	resolved, err := svc.ResolveAccount("BCA")
	if err != nil {
		t.Fatalf("ResolveAccount by name: %v", err)
	}
	if resolved.Name != "BCA" {
		t.Errorf("expected name 'BCA', got '%s'", resolved.Name)
	}
}

func TestResolveAccountNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveAccount("Ghost")
	if err == nil {
		t.Fatal("expected error for unknown account")
	}
}

func TestResolveCategoryWithSuggestions(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveCategory("Restauran")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestDateParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-07-01", "2026-07-01"},
		{"01/07/2026", "2026-07-01"},
		{"01 Jul 2026", "2026-07-01"},
		{"1 Jul 2026", "2026-07-01"},
		{"", ""},
	}

	for _, tc := range tests {
		result, err := parseDate(tc.input)
		if tc.expected == "" {
			if err != nil {
				continue
			}
			if result == "" {
				t.Error("expected today's date for empty input, got empty")
			}
			continue
		}
		if err != nil {
			t.Errorf("parseDate(%q): %v", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("parseDate(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestDateParsingToday(t *testing.T) {
	result, err := parseDate("today")
	if err != nil {
		t.Fatalf("parseDate(today): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'today'")
	}
}

func TestDateParsingYesterday(t *testing.T) {
	result, err := parseDate("yesterday")
	if err != nil {
		t.Fatalf("parseDate(yesterday): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'yesterday'")
	}
}

func TestDateParsingInvalid(t *testing.T) {
	_, err := parseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestMonthParsing(t *testing.T) {
	_, _, err := parseMonth("july")
	if err != nil {
		t.Fatalf("parseMonth(july): %v", err)
	}

	_, _, err = parseMonth("jan")
	if err != nil {
		t.Fatalf("parseMonth(jan): %v", err)
	}

	_, _, err = parseMonth("2026-07")
	if err != nil {
		t.Fatalf("parseMonth(2026-07): %v", err)
	}
}

func TestMonthParsingInvalid(t *testing.T) {
	_, _, err := parseMonth("not-a-month")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestMonthParsingSlashFormat(t *testing.T) {
	from, to, err := parseMonth("07/2026")
	if err != nil {
		t.Fatalf("parseMonth(07/2026): %v", err)
	}
	if from != "2026-07-01" {
		t.Errorf("expected from '2026-07-01', got '%s'", from)
	}
	if to != "2026-07-31" {
		t.Errorf("expected to '2026-07-31', got '%s'", to)
	}
}

func TestParseDateTomorrow(t *testing.T) {
	result, err := parseDate("tomorrow")
	if err != nil {
		t.Fatalf("parseDate(tomorrow): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'tomorrow'")
	}
}

func TestNotFoundErrorUnwrap(t *testing.T) {
	e := &NotFoundError{Entity: "account", Name: "test"}
	if !errors.Is(e, ErrNotFound) {
		t.Error("expected NotFoundError to unwrap to ErrNotFound")
	}
	if errors.Unwrap(e) != ErrNotFound {
		t.Error("expected Unwrap to return ErrNotFound")
	}
}

func TestValidationErrorError(t *testing.T) {
	e := &ValidationError{Field: "name", Message: "cannot be empty"}
	if e.Error() != "name: cannot be empty" {
		t.Errorf("expected 'name: cannot be empty', got '%s'", e.Error())
	}
}

func TestServiceDB(t *testing.T) {
	svc := setupService(t)
	db := svc.DB()
	if db == nil {
		t.Error("expected non-nil DB")
	}
	if err := db.Ping(); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}

func TestServiceQueries(t *testing.T) {
	svc := setupService(t)
	q := svc.Queries()
	if q == nil {
		t.Error("expected non-nil Queries")
	}
}

func TestStringToInterface(t *testing.T) {
	if stringToInterface("") != nil {
		t.Error("expected nil for empty string")
	}
	if stringToInterface("hello") != "hello" {
		t.Error("expected string for non-empty string")
	}
}

func TestSumTransactionAmounts(t *testing.T) {
	svc := setupService(t)

	if total := svc.sumTransactionAmounts(nil); total != 0 {
		t.Errorf("expected 0 for nil, got %d", total)
	}
	if total := svc.sumTransactionAmounts([]*gen.Transaction{}); total != 0 {
		t.Errorf("expected 0 for empty, got %d", total)
	}

	txns := []*gen.Transaction{
		{Amount: 100},
		{Amount: 200},
		{Amount: -50},
	}
	if total := svc.sumTransactionAmounts(txns); total != 250 {
		t.Errorf("expected 250, got %d", total)
	}
}

func TestBalanceToInt64(t *testing.T) {
	if b := balanceToInt64(int64(42)); b != 42 {
		t.Errorf("expected 42, got %d", b)
	}
	if b := balanceToInt64(float64(3.14)); b != 0 {
		t.Errorf("expected 0 for non-int64, got %d", b)
	}
	if b := balanceToInt64("string"); b != 0 {
		t.Errorf("expected 0 for string, got %d", b)
	}
	if b := balanceToInt64(nil); b != 0 {
		t.Errorf("expected 0 for nil, got %d", b)
	}
}

func TestGetAccountByName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetAccountByName("NotExist")
	if err == nil {
		t.Fatal("expected error for non-existent account name")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	account, err := svc.GetAccountByName("BCA")
	if err != nil {
		t.Fatalf("GetAccountByName: %v", err)
	}
	if account.Name != "BCA" {
		t.Errorf("expected name 'BCA', got '%s'", account.Name)
	}
}

func TestListAccounts(t *testing.T) {
	svc := setupService(t)

	accounts, err := svc.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	accounts, err = svc.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestGetAccountByIDNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetAccountByID(9999)
	if err == nil {
		t.Fatal("expected error for non-existent account ID")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestUpdateAccount(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")
	updated, err := svc.UpdateAccount(account.ID, "BCA New", "savings", "USD")
	if err != nil {
		t.Fatalf("UpdateAccount: %v", err)
	}
	if updated.Name != "BCA New" {
		t.Errorf("expected name 'BCA New', got '%s'", updated.Name)
	}
	if updated.Type != "savings" {
		t.Errorf("expected type 'savings', got '%s'", updated.Type)
	}
	if updated.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", updated.Currency)
	}
}

func TestUpdateAccountNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.UpdateAccount(9999, "Ghost", "", "")
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestUpdateAccountEmptyName(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("BCA", "checking", "IDR")
	_, err := svc.UpdateAccount(account.ID, "", "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestArchiveAccount(t *testing.T) {
	svc := setupService(t)

	account, _ := svc.CreateAccount("TempAccount", "checking", "IDR")
	err := svc.ArchiveAccount(account.ID)
	if err != nil {
		t.Fatalf("ArchiveAccount: %v", err)
	}

	active, _ := svc.ListAccounts()
	for _, a := range active {
		if a.ID == account.ID {
			t.Error("expected archived account to not appear in ListAccounts")
		}
	}
}

func TestArchiveAccountNotFound(t *testing.T) {
	svc := setupService(t)

	err := svc.ArchiveAccount(9999)
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateAccountEmptyName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CreateAccount("", "", "")
	if err == nil {
		t.Fatal("expected error for empty account name")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestCreateCategoryEmptyName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CreateCategory("", "", "")
	if err == nil {
		t.Fatal("expected error for empty category name")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestCreateCategoryInvalidParentID(t *testing.T) {
	svc := setupService(t)

	_, err := svc.CreateCategory("Foo", "not-a-number", "")
	if err == nil {
		t.Fatal("expected error for invalid parent ID")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateCategoryEmptyName(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("Old", "", "")
	_, err := svc.UpdateCategory(cat.ID, "", "")
	if err == nil {
		t.Fatal("expected error for empty category name in update")
	}
}

func TestGetCategoryByID(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("Custom", "", "")
	result, err := svc.GetCategoryByID(cat.ID)
	if err != nil {
		t.Fatalf("GetCategoryByID: %v", err)
	}
	if result.Name != "Custom" {
		t.Errorf("expected name 'Custom', got '%s'", result.Name)
	}
}

func TestGetCategoryByIDNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetCategoryByID(9999)
	if err == nil {
		t.Fatal("expected error for non-existent category ID")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestListAllCategories(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("TempCategory", "", "")
	_ = svc.ArchiveCategory(cat.ID)

	categories, err := svc.ListAllCategories()
	if err != nil {
		t.Fatalf("ListAllCategories: %v", err)
	}

	hasArchived := false
	for _, c := range categories {
		if c.ID == cat.ID {
			hasArchived = true
			break
		}
	}
	if !hasArchived {
		t.Error("expected ListAllCategories to include archived categories")
	}
}

func TestGetTagByID(t *testing.T) {
	svc := setupService(t)

	tag, _ := svc.CreateTag("test-tag")
	result, err := svc.GetTagByID(tag.ID)
	if err != nil {
		t.Fatalf("GetTagByID: %v", err)
	}
	if result.Name != "test-tag" {
		t.Errorf("expected name 'test-tag', got '%s'", result.Name)
	}
}

func TestGetTagByIDNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetTagByID(9999)
	if err == nil {
		t.Fatal("expected error for non-existent tag ID")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetTagByName(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateTag("work")
	result, err := svc.GetTagByName("work")
	if err != nil {
		t.Fatalf("GetTagByName: %v", err)
	}
	if result.Name != "work" {
		t.Errorf("expected name 'work', got '%s'", result.Name)
	}
}

func TestGetTagByNameNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetTagByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent tag name")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveTagByIDNotFoundFallsToName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveTag("9999")
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveTagByNameNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveTag("nonexistent-tag")
	if err == nil {
		t.Fatal("expected error for non-existent tag name")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveAccountByIDNotFoundFallsToName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveAccount("9999")
	if err == nil {
		t.Fatal("expected error for non-existent identifier")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveCategoryByIDNotFoundFallsToName(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ResolveCategory("9999")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestGetTransactionByID(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.GetTransactionByID(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("GetTransactionByID: %v", err)
	}
	if result.Amount != 35000 {
		t.Errorf("expected amount 35000, got %d", result.Amount)
	}
}

func TestGetTransactionByIDNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetTransactionByID(9999)
	if err == nil {
		t.Fatal("expected error for non-existent transaction ID")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestListTransactionsEmpty(t *testing.T) {
	svc := setupService(t)

	result, err := svc.ListTransactions(ListTransactionsParams{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(result.Transactions))
	}
}

func TestListTransactionsDefaultDateRange(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithAccount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "BCA expense",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "GoPay expense",
		Category:    "Restaurant",
		Account:     "GoPay",
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		AccountName: "BCA",
	})
	if err != nil {
		t.Fatalf("ListTransactions with account: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 BCA transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithCategory(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "Taxi",
		Category:    "Transportation",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		CategoryName: "Transportation",
	})
	if err != nil {
		t.Fatalf("ListTransactions with category: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 Transportation transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithType(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		Type: "income",
	})
	if err != nil {
		t.Fatalf("ListTransactions with type: %v", err)
	}
	for _, txn := range result.Transactions {
		if txn.Type != "income" {
			t.Errorf("expected only income transactions, got type %s", txn.Type)
		}
	}
}

func TestListTransactionsWithTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("food")
	_, _ = svc.CreateTag("transport")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"food"},
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "Taxi",
		Category:    "Transportation",
		Account:     "BCA",
		Tags:        []string{"transport"},
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		TagName: "food",
	})
	if err != nil {
		t.Fatalf("ListTransactions with tag: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 food-tagged transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithMonth(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "June expense",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-06-15",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		Month: "july",
	})
	if err != nil {
		t.Fatalf("ListTransactions with month: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 July transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithDateRange(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "July 1",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "July 5",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-05",
	})
	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      25000,
		Description: "July 10",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-10",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		DateFrom: "2026-07-01",
		DateTo:   "2026-07-05",
	})
	if err != nil {
		t.Fatalf("ListTransactions with date range: %v", err)
	}
	if len(result.Transactions) != 2 {
		t.Errorf("expected 2 transactions in date range, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithMonthAndTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("food")

	_, _ = svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch July",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"food"},
		Date:        "2026-07-01",
	})

	result, err := svc.ListTransactions(ListTransactionsParams{
		Month:   "july",
		TagName: "food",
	})
	if err != nil {
		t.Fatalf("ListTransactions with month and tag: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Transactions))
	}
}

func TestAddExpenseInvalidDate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "not-a-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestAddIncomeInvalidAmount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      0,
		Description: "Free money",
		Category:    "Salary",
		Account:     "BCA",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = svc.AddIncome(CreateIncomeParams{
		Amount:      -100,
		Description: "Negative income",
		Category:    "Salary",
		Account:     "BCA",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestAddIncomeInvalidDate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Date:        "bad-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestAddIncomeMissingAccount(t *testing.T) {
	svc := setupService(t)

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for missing account")
	}
}

func TestAddIncomeMissingCategory(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "GhostCategory",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestAddTransferInvalidAmount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      0,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = svc.AddTransfer(CreateTransferParams{
		Amount:      -100,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestAddTransferMissingSource(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "Ghost",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error for missing source account")
	}
}

func TestAddTransferMissingDestination(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "Ghost",
	})
	if err == nil {
		t.Fatal("expected error for missing destination account")
	}
}

func TestEditTransactionInvalidAmount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	zeroAmt := int64(0)
	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &zeroAmt,
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for zero, got %v", err)
	}

	negAmt := int64(-100)
	_, err = svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &negAmt,
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestEditTransactionInvalidDate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Date: "not-a-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestEditTransactionMissingAccount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AccountName: "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestEditTransactionMissingCategory(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		CategoryName: "GhostCategory",
	})
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestEditTransactionAddNonexistentTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AddTagNames: []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
}

func TestEditTransactionRemoveNonexistentTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		RemoveTagNames: []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
}

func TestEditTransactionChangeAccount(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	newAmount := int64(40000)
	result, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount:      &newAmount,
		AccountName: "GoPay",
	})
	if err != nil {
		t.Fatalf("EditTransaction change account: %v", err)
	}

	if result.Transaction.Amount != 40000 {
		t.Errorf("expected amount 40000, got %d", result.Transaction.Amount)
	}

	bca, _ := svc.GetAccountByID(1)
	if bca.Balance != 0 {
		t.Errorf("expected BCA balance 0, got %d", bca.Balance)
	}
	gopay, _ := svc.GetAccountByID(2)
	if gopay.Balance != -40000 {
		t.Errorf("expected GoPay balance -40000, got %d", gopay.Balance)
	}
}

func TestEditTransactionTransferRecalculate(t *testing.T) {
	svc := setupService(t)

	acct1, _ := svc.CreateAccount("BCA", "checking", "IDR")
	acct2, _ := svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "2026-07-01",
	})

	newAmount := int64(300000)
	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("EditTransaction transfer: %v", err)
	}

	updatedSrc, _ := svc.GetAccountByID(acct1.ID)
	if updatedSrc.Balance != 700000 {
		t.Errorf("expected source balance 700000, got %d", updatedSrc.Balance)
	}
	updatedDst, _ := svc.GetAccountByID(acct2.ID)
	if updatedDst.Balance != 300000 {
		t.Errorf("expected destination balance 300000, got %d", updatedDst.Balance)
	}
}

func TestRemoveTransfer(t *testing.T) {
	svc := setupService(t)

	acct1, _ := svc.CreateAccount("BCA", "checking", "IDR")
	acct2, _ := svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "2026-07-01",
	})

	err := svc.RemoveTransaction(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("RemoveTransfer: %v", err)
	}

	src, _ := svc.GetAccountByID(acct1.ID)
	if src.Balance != 1000000 {
		t.Errorf("expected source balance 1000000 after removing transfer, got %d", src.Balance)
	}
	dst, _ := svc.GetAccountByID(acct2.ID)
	if dst.Balance != 0 {
		t.Errorf("expected destination balance 0 after removing transfer, got %d", dst.Balance)
	}
}

func TestAdjustBalanceMissingAccount(t *testing.T) {
	svc := setupService(t)

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account: "Ghost",
		Target:  1000000,
	})
	if err == nil {
		t.Fatal("expected error for missing account")
	}
}

func TestEditTransactionWithDateDescriptionNotes(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Date:        "2026-07-15",
		Description: "Updated lunch",
		Notes:       "With friends",
	})
	if err != nil {
		t.Fatalf("EditTransaction: %v", err)
	}
	if result.Transaction.Date != "2026-07-15" {
		t.Errorf("expected date '2026-07-15', got '%s'", result.Transaction.Date)
	}
	if !result.Transaction.Description.Valid || result.Transaction.Description.String != "Updated lunch" {
		t.Errorf("expected description 'Updated lunch', got '%v'", result.Transaction.Description)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "With friends" {
		t.Errorf("expected notes 'With friends', got '%v'", result.Transaction.Notes)
	}
}

func TestResolveTagByID(t *testing.T) {
	svc := setupService(t)

	tag, _ := svc.CreateTag("test-tag")
	resolved, err := svc.ResolveTag(fmt.Sprintf("%d", tag.ID))
	if err != nil {
		t.Fatalf("ResolveTag by ID: %v", err)
	}
	if resolved.ID != tag.ID {
		t.Errorf("expected ID %d, got %d", tag.ID, resolved.ID)
	}
}

func TestResolveCategoryByID(t *testing.T) {
	svc := setupService(t)

	cat, _ := svc.CreateCategory("CustomCat", "", "")
	resolved, err := svc.ResolveCategory(fmt.Sprintf("%d", cat.ID))
	if err != nil {
		t.Fatalf("ResolveCategory by ID: %v", err)
	}
	if resolved.ID != cat.ID {
		t.Errorf("expected ID %d, got %d", cat.ID, resolved.ID)
	}
}

func TestAddTransferInvalidDate(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "invalid-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid date in transfer")
	}
}

func TestAddIncomeMissingTag(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Tags:        []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestListTransactionsInvalidMonth(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ListTransactions(ListTransactionsParams{
		Month: "smarch",
	})
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestListTransactionsInvalidAccount(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ListTransactions(ListTransactionsParams{
		AccountName: "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestListTransactionsInvalidCategory(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ListTransactions(ListTransactionsParams{
		CategoryName: "GhostCategory",
	})
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestAddIncomeWithTags(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateTag("salary-tag")
	_, _ = svc.CreateTag("bonus")

	result, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Monthly salary",
		Category:    "Salary",
		Account:     "BCA",
		Tags:        []string{"salary-tag", "bonus"},
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddIncome with tags: %v", err)
	}

	tags, err := svc.ListTransactionTags(result.Transaction.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestGetAccountByID_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetAccountByID(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("expected non-NotFoundError, got %v", err)
	}
}

func TestGetAccountByName_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetAccountByName("BCA")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestResolveAccount_DBErrorByID(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveAccount("1")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestResolveAccount_DBErrorByName(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveAccount("BCA")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestArchiveAccount_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	err := svc.ArchiveAccount(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestCreateAccount_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.CreateAccount("Test", "checking", "IDR")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestUpdateAccount_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.UpdateAccount(1, "New", "", "")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestGetCategoryByID_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetCategoryByID(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestResolveCategory_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveCategory("1")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestCreateCategory_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.CreateCategory("Test", "", "")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestUpdateCategory_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.UpdateCategory(1, "New", "")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestArchiveCategory_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	err := svc.ArchiveCategory(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestGetTagByID_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetTagByID(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestGetTagByName_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetTagByName("work")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestResolveTag_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveTag("1")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestDeleteTag_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	err := svc.DeleteTag(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestGetTransactionByID_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.GetTransactionByID(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestRecalculateBalanceDBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	err := svc.recalculateBalance(1)
	if err == nil {
		t.Fatal("expected error from recalculateBalance with closed DB")
	}
}

func TestAddExpense_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_ = svc.DB().Close()

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestAddIncome_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_ = svc.DB().Close()

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestAddTransfer_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")
	_ = svc.DB().Close()

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestListTransactions_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ListTransactions(ListTransactionsParams{
		Limit: 10,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestListTransactionsByTag_DBError(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ListTransactions(ListTransactionsParams{
		TagName: "food",
		Limit:   10,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestEditTransaction_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_ = svc.DB().Close()

	_, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestRemoveTransaction_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_ = svc.DB().Close()

	err := svc.RemoveTransaction(txn.Transaction.ID)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestAdjustBalance_DBError(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})

	_ = svc.DB().Close()

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account: "BCA",
		Target:  2000000,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestCreateAccountDefaultType(t *testing.T) {
	svc := setupService(t)

	account, err := svc.CreateAccount("DefaultType", "", "IDR")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if account.Type != "checking" {
		t.Errorf("expected type 'checking', got '%s'", account.Type)
	}
}

func TestCreateAccountDefaultCurrency(t *testing.T) {
	svc := setupService(t)

	account, err := svc.CreateAccount("DefaultCurr", "checking", "")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if account.Currency != "IDR" {
		t.Errorf("expected currency 'IDR', got '%s'", account.Currency)
	}
}

func TestCreateAccountDefaultTypeAndCurrency(t *testing.T) {
	svc := setupService(t)

	account, err := svc.CreateAccount("AllDefaults", "", "")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if account.Type != "checking" {
		t.Errorf("expected type 'checking', got '%s'", account.Type)
	}
	if account.Currency != "IDR" {
		t.Errorf("expected currency 'IDR', got '%s'", account.Currency)
	}
}

func TestAddIncomeWithNotes(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	result, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Notes:       "Monthly salary for July",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddIncome with notes: %v", err)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "Monthly salary for July" {
		t.Errorf("expected notes 'Monthly salary for July', got '%v'", result.Transaction.Notes)
	}
}

func TestAddTransferWithDescriptionAndNotes(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Description: "Transfer to e-wallet",
		Notes:       "For online shopping",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddTransfer with description and notes: %v", err)
	}
	if !result.Transaction.Description.Valid || result.Transaction.Description.String != "Transfer to e-wallet" {
		t.Errorf("expected description 'Transfer to e-wallet', got '%v'", result.Transaction.Description)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "For online shopping" {
		t.Errorf("expected notes 'For online shopping', got '%v'", result.Transaction.Notes)
	}
}

func TestEditTransactionWithCategoryName(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	txn, _ := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := svc.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		CategoryName: "Transportation",
	})
	if err != nil {
		t.Fatalf("EditTransaction with category: %v", err)
	}
	if !result.Transaction.CategoryID.Valid {
		t.Error("expected category_id to be set")
	}

	cat, _ := svc.ResolveCategory("Transportation")
	if result.Transaction.CategoryID.Int64 != cat.ID {
		t.Errorf("expected category ID %d, got %d", cat.ID, result.Transaction.CategoryID.Int64)
	}
}

func TestAddExpenseWithNotes(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Notes:       "Lunch with colleagues",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense with notes: %v", err)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "Lunch with colleagues" {
		t.Errorf("expected notes 'Lunch with colleagues', got '%v'", result.Transaction.Notes)
	}
}

func TestResolveTag_DBErrorByName(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveTag("work")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestAdjustBalanceWithNotes(t *testing.T) {
	svc := setupService(t)

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      2000000,
		Description: "Balance correction",
		Notes:       "Adjusted after review",
	})
	if err != nil {
		t.Fatalf("AdjustBalance with notes: %v", err)
	}
	if result.Transaction.Notes.Valid && result.Transaction.Notes.String == "Adjusted after review" {
		return
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "Adjusted after review" {
		t.Errorf("expected notes 'Adjusted after review', got '%v'", result.Transaction.Notes)
	}
}

func TestCreateCategoryParentLookupDBError(t *testing.T) {
	svc := setupService(t)

	parent, _ := svc.CreateCategory("Parent", "", "")
	parentIDStr := fmt.Sprintf("%d", parent.ID)
	_ = svc.DB().Close()

	_, err := svc.CreateCategory("Child", parentIDStr, "")
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestNewWithQuerier(t *testing.T) {
	dbase := testdb.Open(t)
	svc := NewWithQuerier(dbase, gen.New(dbase))
	if svc == nil {
		t.Fatal("expected non-nil Service")
	}
	if svc.q == nil {
		t.Fatal("expected non-nil querier")
	}
}

type createFailQuerier struct {
	gen.Querier
}

func (c createFailQuerier) CreateTransaction(ctx context.Context, arg gen.CreateTransactionParams) (*gen.Transaction, error) {
	return nil, fmt.Errorf("mock create failure")
}

func TestAddExpenseCreateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	_, _ = gen.New(dbase).CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, createFailQuerier{gen.New(dbase)})

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestAddIncomeCreateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	_, _ = gen.New(dbase).CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, createFailQuerier{gen.New(dbase)})

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestAddTransferCreateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	svc := NewWithQuerier(dbase, createFailQuerier{q})

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

type updateFailQuerier struct {
	gen.Querier
}

func (u updateFailQuerier) UpdateTransaction(ctx context.Context, arg gen.UpdateTransactionParams) (*gen.Transaction, error) {
	return nil, fmt.Errorf("mock update failure")
}

func TestEditTransactionUpdateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, updateFailQuerier{q})

	_, err := svc.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

type archiveFailQuerier struct {
	gen.Querier
}

func (a archiveFailQuerier) ArchiveTransaction(ctx context.Context, id int64) error {
	return fmt.Errorf("mock archive failure")
}

func TestRemoveTransactionArchiveFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, archiveFailQuerier{q})

	err := svc.RemoveTransaction(txn.ID)
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

type balanceFailQuerier struct {
	gen.Querier
}

func (b balanceFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	return nil, fmt.Errorf("mock balance failure")
}

func TestAdjustBalanceBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, balanceFailQuerier{q})

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account: "BCA",
		Target:  1000000,
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

type tagAddFailQuerier struct {
	gen.Querier
}

func (t tagAddFailQuerier) AddTransactionTag(ctx context.Context, arg gen.AddTransactionTagParams) error {
	return fmt.Errorf("mock tag add failure")
}

type tagRemoveFailQuerier struct {
	gen.Querier
}

func (t tagRemoveFailQuerier) RemoveTransactionTag(ctx context.Context, arg gen.RemoveTransactionTagParams) error {
	return fmt.Errorf("mock tag remove failure")
}

func TestAddExpenseTagAddFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "lunch")
	svc := NewWithQuerier(dbase, tagAddFailQuerier{q})

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"lunch"},
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error from tag add failure")
	}
}

func TestAddIncomeTagAddFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "salary-tag")
	svc := NewWithQuerier(dbase, tagAddFailQuerier{q})

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Tags:        []string{"salary-tag"},
	})
	if err == nil {
		t.Fatal("expected error from tag add failure")
	}
}

type balanceTypeQuerier struct {
	gen.Querier
}

func (b balanceTypeQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	return "not-an-int64", nil
}

func TestRecalculateBalanceTypeError(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, balanceTypeQuerier{q})

	err := svc.recalculateBalance(txn.AccountID)
	if err == nil {
		t.Fatal("expected type error from recalculateBalance")
	}
	if !strings.Contains(err.Error(), "unexpected balance type") {
		t.Errorf("expected 'unexpected balance type' error, got: %v", err)
	}
}

func TestAddExpenseBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, balanceTypeQuerier{q})

	_, err := svc.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected balance recalc error")
	}
}

func TestAddIncomeBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, balanceTypeQuerier{q})

	_, err := svc.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected balance recalc error")
	}
}

type transferBalanceFailQuerier struct {
	gen.Querier
	calls int
}

func (t *transferBalanceFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	t.calls++
	if t.calls > 1 {
		return nil, fmt.Errorf("mock recalc balance failure")
	}
	return int64(0), nil
}

func TestAddTransferRecalcBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	svc := NewWithQuerier(dbase, &transferBalanceFailQuerier{Querier: q})

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected recalc balance failure during transfer")
	}
}

func TestEditTransactionBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, balanceFailQuerier{q})

	_, err := svc.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected balance recalc error during edit")
	}
}

func TestRemoveTransactionRecalcBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, balanceFailQuerier{q})

	err := svc.RemoveTransaction(txn.ID)
	if err == nil {
		t.Fatal("expected balance recalc error during remove")
	}
}

func TestAdjustBalanceCreateTransactionFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, createFailQuerier{q})

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1000000,
		Description: "Adjustment",
	})
	if err == nil {
		t.Fatal("expected create failure during adjust")
	}
}

func TestEditTransactionTagAddFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "test-tag")
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, tagAddFailQuerier{q})

	_, err := svc.EditTransaction(txn.ID, EditTransactionParams{
		AddTagNames: []string{"test-tag"},
	})
	if err == nil {
		t.Fatal("expected tag add failure during edit")
	}
}

func TestEditTransactionTagRemoveFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "test-tag")
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	_ = q.AddTransactionTag(context.Background(), gen.AddTransactionTagParams{TransactionID: txn.ID, TagID: 1})

	svc := NewWithQuerier(dbase, tagRemoveFailQuerier{q})

	_, err := svc.EditTransaction(txn.ID, EditTransactionParams{
		RemoveTagNames: []string{"test-tag"},
	})
	if err == nil {
		t.Fatal("expected tag remove failure during edit")
	}
}

type listTagsFailQuerier struct {
	gen.Querier
}

func (l listTagsFailQuerier) ListTransactionTags(ctx context.Context, transactionID int64) ([]*gen.Tag, error) {
	return nil, fmt.Errorf("mock list tags failure")
}

func TestEditTransactionListTagsFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	svc := NewWithQuerier(dbase, listTagsFailQuerier{q})

	_, err := svc.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected list tags failure during edit")
	}
}

type transferDestFailQuerier struct {
	gen.Querier
	calls int
}

func (t *transferDestFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	t.calls++
	if t.calls >= 3 {
		return nil, fmt.Errorf("mock destination recalc failure")
	}
	return int64(0), nil
}

func TestAddTransferDestinationRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	svc := NewWithQuerier(dbase, &transferDestFailQuerier{Querier: q})

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected destination recalc failure during transfer")
	}
}

type adjustRecalcFailQuerier struct {
	gen.Querier
	callCount int
}

func (a *adjustRecalcFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	a.callCount++
	if a.callCount == 2 {
		return nil, fmt.Errorf("mock adjust recalc failure")
	}
	return int64(0), nil
}

func TestAdjustBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, &adjustRecalcFailQuerier{Querier: q})

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1000000,
		Description: "Adjustment",
	})
	if err == nil {
		t.Fatal("expected adjust recalc failure")
	}
}

func TestAddTransferDescriptionOnly(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Description: "Transfer desc",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddTransfer desc only: %v", err)
	}
	if !result.Transaction.Description.Valid || result.Transaction.Description.String != "Transfer desc" {
		t.Errorf("expected description 'Transfer desc'")
	}
}

func TestAddTransferNotesOnly(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := svc.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Notes:       "Transfer notes only",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddTransfer notes only: %v", err)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "Transfer notes only" {
		t.Errorf("expected notes 'Transfer notes only'")
	}
}

func TestAdjustBalanceNotesOnly(t *testing.T) {
	svc := setupService(t)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := svc.AdjustBalance(AdjustBalanceParams{
		Account: "BCA",
		Target:  2000000,
		Notes:   "Adjust notes only",
	})
	if err != nil {
		t.Fatalf("AdjustBalance notes only: %v", err)
	}
	if !result.Transaction.Notes.Valid || result.Transaction.Notes.String != "Adjust notes only" {
		t.Errorf("expected notes 'Adjust notes only'")
	}
}

type adjustFinalFailQuerier struct {
	gen.Querier
	callCount int
}

func (a *adjustFinalFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	a.callCount++
	if a.callCount == 3 {
		return nil, fmt.Errorf("mock final balance failure")
	}
	return int64(0), nil
}

type transferFirstBalanceFailQuerier struct {
	gen.Querier
}

func (t transferFirstBalanceFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	return nil, fmt.Errorf("mock balance check failure")
}

func TestAddTransferGetBalanceError(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	svc := NewWithQuerier(dbase, transferFirstBalanceFailQuerier{q})

	_, err := svc.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected get balance error")
	}
}

func TestAdjustBalanceGetNewBalanceError(t *testing.T) {
	dbase := testdb.Open(t)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	svc := NewWithQuerier(dbase, &adjustFinalFailQuerier{Querier: q})

	_, err := svc.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1000000,
		Description: "Adjustment",
	})
	if err == nil {
		t.Fatal("expected final balance read error")
	}
}

func setupServiceWithMultiCurrency(t *testing.T) *Service {
	t.Helper()
	svc := setupService(t)

	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})
	t.Cleanup(ResetTestRateConfig)

	return svc
}

func TestAddExpenseForeignCurrency(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	account, err := svc.CreateAccount("Wise USD", "checking", "USD")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      10,
		Description: "AWS monthly",
		Category:    "Subscriptions",
		Account:     "Wise USD",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	if result.Transaction.Amount != 10 {
		t.Errorf("expected original amount 10, got %d", result.Transaction.Amount)
	}
	if result.Transaction.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", result.Transaction.Currency)
	}

	if !result.Transaction.BaseAmount.Valid {
		t.Fatal("expected base_amount to be set for foreign currency")
	}
	expectedBase := int64(10 * 15800)
	if result.Transaction.BaseAmount.Int64 != expectedBase {
		t.Errorf("expected base_amount %d, got %d", expectedBase, result.Transaction.BaseAmount.Int64)
	}
	if !result.Transaction.BaseCurrency.Valid {
		t.Fatal("expected base_currency to be set")
	}
	if result.Transaction.BaseCurrency.String != "IDR" {
		t.Errorf("expected base_currency IDR, got %s", result.Transaction.BaseCurrency.String)
	}

	updated, err := svc.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -10 {
		t.Errorf("expected balance -10 (in account currency USD), got %d", updated.Balance)
	}
}

func TestAddIncomeForeignCurrency(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	account, err := svc.CreateAccount("PayPal USD", "checking", "USD")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.AddIncome(CreateIncomeParams{
		Amount:      100,
		Description: "Freelance payment",
		Category:    "Freelance",
		Account:     "PayPal USD",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddIncome: %v", err)
	}

	if result.Transaction.Amount != 100 {
		t.Errorf("expected original amount 100, got %d", result.Transaction.Amount)
	}
	if result.Transaction.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", result.Transaction.Currency)
	}

	if !result.Transaction.BaseAmount.Valid {
		t.Fatal("expected base_amount to be set for foreign currency")
	}
	expectedBase := int64(100 * 15800)
	if result.Transaction.BaseAmount.Int64 != expectedBase {
		t.Errorf("expected base_amount %d, got %d", expectedBase, result.Transaction.BaseAmount.Int64)
	}

	updated, err := svc.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 100 {
		t.Errorf("expected balance 100 (in account currency USD), got %d", updated.Balance)
	}
}

func TestAddExpenseBaseCurrency(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	account, err := svc.CreateAccount("GoPay", "ewallet", "IDR")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Coffee",
		Category:    "Coffee & Snacks",
		Account:     "GoPay",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	if result.Transaction.Currency != "IDR" {
		t.Errorf("expected currency IDR, got %s", result.Transaction.Currency)
	}
	if result.Transaction.BaseAmount.Valid {
		t.Errorf("expected base_amount to be unset for base currency, got %d", result.Transaction.BaseAmount.Int64)
	}
	if result.Transaction.BaseCurrency.Valid {
		t.Errorf("expected base_currency to be unset for base currency, got %s", result.Transaction.BaseCurrency.String)
	}

	updated, err := svc.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -50000 {
		t.Errorf("expected balance -50000, got %d", updated.Balance)
	}
}

func TestAddExpenseForeignCurrencyEUR(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	account, err := svc.CreateAccount("Revolut EUR", "checking", "EUR")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	result, err := svc.AddExpense(CreateExpenseParams{
		Amount:      50,
		Description: "Hotel booking",
		Category:    "Travel",
		Account:     "Revolut EUR",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	if result.Transaction.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %s", result.Transaction.Currency)
	}

	if !result.Transaction.BaseAmount.Valid {
		t.Fatal("expected base_amount to be set")
	}
	expectedBase := int64(50 * 17200)
	if result.Transaction.BaseAmount.Int64 != expectedBase {
		t.Errorf("expected base_amount %d, got %d", expectedBase, result.Transaction.BaseAmount.Int64)
	}

	updated, err := svc.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -50 {
		t.Errorf("expected balance -50 (in EUR), got %d", updated.Balance)
	}
}

func TestAddExpenseMissingRate(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	_, err := svc.CreateAccount("KRW Account", "checking", "KRW")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Korean food",
		Category:    "Restaurant",
		Account:     "KRW Account",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error for missing rate")
	}

	var rnf *RateNotFoundError
	if !errors.As(err, &rnf) {
		t.Errorf("expected RateNotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "wallet rate add KRW") {
		t.Errorf("expected actionable error, got: %v", err)
	}

	transactions, err := svc.ListTransactions(ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(transactions.Transactions) > 0 {
		t.Errorf("expected no transaction to be persisted, got %d", len(transactions.Transactions))
	}
}

func TestAddIncomeMissingRate(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	_, err := svc.CreateAccount("JPY Account", "checking", "JPY")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = svc.AddIncome(CreateIncomeParams{
		Amount:      5000,
		Description: "Japanese income",
		Category:    "Salary",
		Account:     "JPY Account",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error for missing rate")
	}

	transactions, err := svc.ListTransactions(ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(transactions.Transactions) > 0 {
		t.Errorf("expected no transaction to be persisted, got %d", len(transactions.Transactions))
	}
}

func TestListTransactionsWithOnlyBaseCurrency(t *testing.T) {
	svc := setupServiceWithMultiCurrency(t)

	_, err := svc.CreateAccount("BCA", "checking", "IDR")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = svc.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Coffee",
		Category:    "Coffee & Snacks",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	result, err := svc.ListTransactions(ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}

	if result.BaseTotal != 0 {
		t.Errorf("expected BaseTotal 0 for base-only transactions, got %d", result.BaseTotal)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Transactions))
	}
}

func TestResolveBaseFieldsMissingRateConfig(t *testing.T) {
	origLoad := svcLoadRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, ErrRateConfigMissing
	}
	defer func() { svcLoadRates = origLoad }()

	svc := New(testdb.Open(t))
	_, _, err := svc.resolveBaseFields("USD", 100)
	if err == nil {
		t.Fatal("expected error for missing rate config in resolveBaseFields")
	}
}
