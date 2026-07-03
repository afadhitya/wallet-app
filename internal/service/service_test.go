package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/afadhitya/wallet-app/internal/testdb"
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
