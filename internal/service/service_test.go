package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/testdb"
)

func setupService(t *testing.T) *Service {
	t.Helper()
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(ResetTestRateConfig)
	return New(testdb.Open(t))
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

func TestListAllAccounts(t *testing.T) {
	svc := setupService(t)

	accounts, err := svc.ListAllAccounts()
	if err != nil {
		t.Fatalf("ListAllAccounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}

	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	_ = svc.ArchiveAccount(1)

	accounts, err = svc.ListAllAccounts()
	if err != nil {
		t.Fatalf("ListAllAccounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts including archived, got %d", len(accounts))
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
	updated, err := svc.UpdateAccount(account.ID, "BCA New", "savings", "USD", 0)
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

	_, err := svc.UpdateAccount(9999, "Ghost", "", "", 0)
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
	_, err := svc.UpdateAccount(account.ID, "", "", "", 0)
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

	_, err := svc.UpdateAccount(1, "New", "", "", 0)
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

func TestResolveTag_DBErrorByName(t *testing.T) {
	svc := setupService(t)
	_ = svc.DB().Close()

	_, err := svc.ResolveTag("work")
	if err == nil {
		t.Fatal("expected error with closed DB")
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

