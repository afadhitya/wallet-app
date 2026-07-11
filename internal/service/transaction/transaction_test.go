package transaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
	"github.com/afadhitya/wallet-app/internal/testdb"
	"github.com/afadhitya/wallet-app/pkg/config"
)

func setupManager(t *testing.T) *Manager {
	t.Helper()
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	return NewManager(gen.New(testdb.Open(t)))
}

func setupManagerWithMultiCurrency(t *testing.T) *Manager {
	t.Helper()
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	return NewManager(gen.New(testdb.Open(t)))
}

func mustCreateAccount(t *testing.T, m *Manager, name, typ, currency string) *gen.Account {
	t.Helper()
	acc, err := m.q.CreateAccount(context.Background(), gen.CreateAccountParams{
		Name: name, Type: typ, Currency: currency,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return acc
}

func mustCreateTag(t *testing.T, m *Manager, name string) *gen.Tag {
	t.Helper()
	tag, err := m.q.CreateTag(context.Background(), name)
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	return tag
}

func TestAddExpense(t *testing.T) {
	m := setupManager(t)

	account := mustCreateAccount(t, m, "BCA", "checking", "IDR")

	result, err := m.AddExpense(CreateExpenseParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -35000 {
		t.Errorf("expected balance -35000, got %d", updated.Balance)
	}
}

func TestAddExpenseWithTags(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateTag(t, m, "lunch")
	_ = mustCreateTag(t, m, "work")

	result, err := m.AddExpense(CreateExpenseParams{
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

	tags, err := m.q.ListTransactionTags(context.Background(), result.Transaction.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestAddExpenseInvalidAmount(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddExpense(CreateExpenseParams{
		Amount:      0,
		Description: "Free",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = m.AddExpense(CreateExpenseParams{
		Amount:      -100,
		Description: "Negative",
		Category:    "Restaurant",
		Account:     "BCA",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative amount, got %v", err)
	}
}

func TestAddExpenseMissingAccount(t *testing.T) {
	m := setupManager(t)

	_, err := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for missing account")
	}
	var notFound *shared.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestAddExpenseMissingCategory(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddExpense(CreateExpenseParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddExpense(CreateExpenseParams{
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
	m := setupManager(t)

	account := mustCreateAccount(t, m, "BCA", "checking", "IDR")

	result, err := m.AddIncome(CreateIncomeParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 5000000 {
		t.Errorf("expected balance 5000000, got %d", updated.Balance)
	}
}

func TestAddTransfer(t *testing.T) {
	m := setupManager(t)

	src := mustCreateAccount(t, m, "BCA", "checking", "IDR")
	dst := mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := m.AddTransfer(CreateTransferParams{
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

	srcUpdated, err := m.q.GetAccountByID(context.Background(), src.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if srcUpdated.Balance != 800000 {
		t.Errorf("expected source balance 800000, got %d", srcUpdated.Balance)
	}

	dstUpdated, err := m.q.GetAccountByID(context.Background(), dst.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if dstUpdated.Balance != 200000 {
		t.Errorf("expected destination balance 200000, got %d", dstUpdated.Balance)
	}
}

func TestAddTransferInsufficientBalance(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	result, err := m.AddTransfer(CreateTransferParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "BCA",
	})
	if err == nil {
		t.Fatal("expected error for same source/destination")
	}
}

func TestEditTransaction(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	newAmount := int64(40000)
	result, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("EditTransaction: %v", err)
	}

	if result.Transaction.Amount != 40000 {
		t.Errorf("expected amount 40000, got %d", result.Transaction.Amount)
	}

	updated, err := m.q.GetAccountByID(context.Background(), txn.Transaction.AccountID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -40000 {
		t.Errorf("expected balance -40000, got %d", updated.Balance)
	}
}

func TestEditTransactionTags(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateTag(t, m, "lunch")
	_ = mustCreateTag(t, m, "work")

	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"lunch"},
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AddTagNames:    []string{"work"},
		RemoveTagNames: []string{"lunch"},
	})
	if err != nil {
		t.Fatalf("EditTransaction tags: %v", err)
	}

	tags, err := m.q.ListTransactionTags(context.Background(), txn.Transaction.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "work" {
		t.Errorf("expected tag 'work', got '%s'", tags[0].Name)
	}
}

func TestRemoveTransaction(t *testing.T) {
	m := setupManager(t)

	account := mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      5000000,
		Description: "Income",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      1000000,
		Description: "Big expense",
		Category:    "Restaurant",
		Account:     "BCA",
	})

	err := m.RemoveTransaction(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("RemoveTransaction: %v", err)
	}

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 5000000 {
		t.Errorf("expected balance 5000000 after removal, got %d", updated.Balance)
	}
}

func TestRemoveTransactionNotFound(t *testing.T) {
	m := setupManager(t)

	err := m.RemoveTransaction(9999)
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
	var notFound *shared.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestAdjustBalance(t *testing.T) {
	m := setupManager(t)

	account := mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := m.AdjustBalance(AdjustBalanceParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 1500000 {
		t.Errorf("expected balance 1500000, got %d", updated.Balance)
	}
}

func TestBalanceRecalculationIgnoreArchived(t *testing.T) {
	m := setupManager(t)

	account := mustCreateAccount(t, m, "BCA", "checking", "IDR")

	txn1, _ := m.AddIncome(CreateIncomeParams{
		Amount:      5000000,
		Description: "Income",
		Category:    "Salary",
		Account:     "BCA",
	})
	txn2, _ := m.AddIncome(CreateIncomeParams{
		Amount:      3000000,
		Description: "Bonus",
		Category:    "Salary",
		Account:     "BCA",
	})

	_ = m.RemoveTransaction(txn1.Transaction.ID)

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 3000000 {
		t.Errorf("expected balance 3000000 after archiving income, got %d", updated.Balance)
	}
	_ = txn2
}

func TestEditTransactionNotFound(t *testing.T) {
	m := setupManager(t)

	_, err := m.EditTransaction(9999, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
}

func TestAdjustBalanceNoChange(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	result, err := m.AdjustBalance(AdjustBalanceParams{
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

func TestGetTransactionByID(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.GetTransactionByID(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("GetTransactionByID: %v", err)
	}
	if result.Amount != 35000 {
		t.Errorf("expected amount 35000, got %d", result.Amount)
	}
}

func TestGetTransactionByIDNotFound(t *testing.T) {
	m := setupManager(t)

	_, err := m.GetTransactionByID(9999)
	if err == nil {
		t.Fatal("expected error for non-existent transaction ID")
	}
	var notFound *shared.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestListTransactionsEmpty(t *testing.T) {
	m := setupManager(t)

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Transactions))
	}
}

func TestListTransactionsWithAccount(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "BCA expense",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "GoPay expense",
		Category:    "Restaurant",
		Account:     "GoPay",
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "Taxi",
		Category:    "Transportation",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateTag(t, m, "food")
	_ = mustCreateTag(t, m, "transport")

	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"food"},
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "Taxi",
		Category:    "Transportation",
		Account:     "BCA",
		Tags:        []string{"transport"},
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "June expense",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-06-15",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "July 1",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      15000,
		Description: "July 5",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-05",
	})
	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      25000,
		Description: "July 10",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-10",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateTag(t, m, "food")

	_, _ = m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch July",
		Category:    "Restaurant",
		Account:     "BCA",
		Tags:        []string{"food"},
		Date:        "2026-07-01",
	})

	result, err := m.ListTransactions(ListTransactionsParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddExpense(CreateExpenseParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddIncome(CreateIncomeParams{
		Amount:      0,
		Description: "Free money",
		Category:    "Salary",
		Account:     "BCA",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = m.AddIncome(CreateIncomeParams{
		Amount:      -100,
		Description: "Negative income",
		Category:    "Salary",
		Account:     "BCA",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestAddIncomeInvalidDate(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddIncome(CreateIncomeParams{
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
	m := setupManager(t)

	_, err := m.AddIncome(CreateIncomeParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddIncome(CreateIncomeParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      0,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = m.AddTransfer(CreateTransferParams{
		Amount:      -100,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestAddTransferMissingSource(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "Ghost",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error for missing source account")
	}
}

func TestAddTransferMissingDestination(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "Ghost",
	})
	if err == nil {
		t.Fatal("expected error for missing destination account")
	}
}

func TestEditTransactionInvalidAmount(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	zeroAmt := int64(0)
	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &zeroAmt,
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for zero, got %v", err)
	}

	negAmt := int64(-100)
	_, err = m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &negAmt,
	})
	if !errors.Is(err, shared.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestEditTransactionInvalidDate(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Date: "not-a-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestEditTransactionMissingAccount(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AccountName: "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestEditTransactionMissingCategory(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		CategoryName: "GhostCategory",
	})
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestEditTransactionAddNonexistentTag(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		AddTagNames: []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
}

func TestEditTransactionRemoveNonexistentTag(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		RemoveTagNames: []string{"nonexistent"},
	})
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
}

func TestEditTransactionChangeAccount(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	newAmount := int64(40000)
	result, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount:      &newAmount,
		AccountName: "GoPay",
	})
	if err != nil {
		t.Fatalf("EditTransaction change account: %v", err)
	}

	if result.Transaction.Amount != 40000 {
		t.Errorf("expected amount 40000, got %d", result.Transaction.Amount)
	}

	bca, err := m.q.GetAccountByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get BCA: %v", err)
	}
	if bca.Balance != 0 {
		t.Errorf("expected BCA balance 0, got %d", bca.Balance)
	}
	gopay, err := m.q.GetAccountByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("get GoPay: %v", err)
	}
	if gopay.Balance != -40000 {
		t.Errorf("expected GoPay balance -40000, got %d", gopay.Balance)
	}
}

func TestEditTransactionTransferRecalculate(t *testing.T) {
	m := setupManager(t)

	acct1 := mustCreateAccount(t, m, "BCA", "checking", "IDR")
	acct2 := mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := m.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "2026-07-01",
	})

	newAmount := int64(300000)
	_, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("EditTransaction transfer: %v", err)
	}

	updatedSrc, err := m.q.GetAccountByID(context.Background(), acct1.ID)
	if err != nil {
		t.Fatalf("get src: %v", err)
	}
	if updatedSrc.Balance != 700000 {
		t.Errorf("expected source balance 700000, got %d", updatedSrc.Balance)
	}
	updatedDst, err := m.q.GetAccountByID(context.Background(), acct2.ID)
	if err != nil {
		t.Fatalf("get dst: %v", err)
	}
	if updatedDst.Balance != 300000 {
		t.Errorf("expected destination balance 300000, got %d", updatedDst.Balance)
	}
}

func TestRemoveTransfer(t *testing.T) {
	m := setupManager(t)

	acct1 := mustCreateAccount(t, m, "BCA", "checking", "IDR")
	acct2 := mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	txn, _ := m.AddTransfer(CreateTransferParams{
		Amount:      200000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
		Date:        "2026-07-01",
	})

	err := m.RemoveTransaction(txn.Transaction.ID)
	if err != nil {
		t.Fatalf("RemoveTransfer: %v", err)
	}

	src, err := m.q.GetAccountByID(context.Background(), acct1.ID)
	if err != nil {
		t.Fatalf("get src: %v", err)
	}
	if src.Balance != 1000000 {
		t.Errorf("expected source balance 1000000 after removing transfer, got %d", src.Balance)
	}
	dst, err := m.q.GetAccountByID(context.Background(), acct2.ID)
	if err != nil {
		t.Fatalf("get dst: %v", err)
	}
	if dst.Balance != 0 {
		t.Errorf("expected destination balance 0 after removing transfer, got %d", dst.Balance)
	}
}

func TestAdjustBalanceMissingAccount(t *testing.T) {
	m := setupManager(t)

	_, err := m.AdjustBalance(AdjustBalanceParams{
		Account: "Ghost",
		Target:  1000000,
	})
	if err == nil {
		t.Fatal("expected error for missing account")
	}
}

func TestEditTransactionWithDateDescriptionNotes(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
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

func TestAddTransferInvalidDate(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, err := m.AddTransfer(CreateTransferParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddIncome(CreateIncomeParams{
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
	m := setupManager(t)

	_, err := m.ListTransactions(ListTransactionsParams{
		Month: "smarch",
	})
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestListTransactionsInvalidAccount(t *testing.T) {
	m := setupManager(t)

	_, err := m.ListTransactions(ListTransactionsParams{
		AccountName: "GhostAccount",
	})
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestListTransactionsInvalidCategory(t *testing.T) {
	m := setupManager(t)

	_, err := m.ListTransactions(ListTransactionsParams{
		CategoryName: "GhostCategory",
	})
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestAddIncomeWithTags(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateTag(t, m, "salary-tag")
	_ = mustCreateTag(t, m, "bonus")

	result, err := m.AddIncome(CreateIncomeParams{
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

	tags, err := m.q.ListTransactionTags(context.Background(), result.Transaction.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestAddExpenseWithNotes(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	result, err := m.AddExpense(CreateExpenseParams{
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

func TestAddIncomeWithNotes(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	result, err := m.AddIncome(CreateIncomeParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Initial",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := m.AddTransfer(CreateTransferParams{
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
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	result, err := m.EditTransaction(txn.Transaction.ID, EditTransactionParams{
		CategoryName: "Transportation",
	})
	if err != nil {
		t.Fatalf("EditTransaction with category: %v", err)
	}
	if !result.Transaction.CategoryID.Valid {
		t.Error("expected category_id to be set")
	}

	cat, err := shared.ResolveCategory(m.q, "Transportation")
	if err != nil {
		t.Fatalf("resolve category: %v", err)
	}
	if result.Transaction.CategoryID.Int64 != cat.ID {
		t.Errorf("expected category ID %d, got %d", cat.ID, result.Transaction.CategoryID.Int64)
	}
}

func TestAdjustBalanceWithNotes(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})

	result, err := m.AdjustBalance(AdjustBalanceParams{
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

func TestAddTransferDescriptionOnly(t *testing.T) {
	m := setupManager(t)
	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := m.AddTransfer(CreateTransferParams{
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
	m := setupManager(t)
	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_ = mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := m.AddTransfer(CreateTransferParams{
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
	m := setupManager(t)
	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount: 1000000, Description: "Init", Category: "Salary", Account: "BCA",
	})

	result, err := m.AdjustBalance(AdjustBalanceParams{
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

func TestAddExpenseForeignCurrency(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	account := mustCreateAccount(t, m, "Wise USD", "checking", "USD")

	result, err := m.AddExpense(CreateExpenseParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -10 {
		t.Errorf("expected balance -10 (in account currency USD), got %d", updated.Balance)
	}
}

func TestAddIncomeForeignCurrency(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	account := mustCreateAccount(t, m, "PayPal USD", "checking", "USD")

	result, err := m.AddIncome(CreateIncomeParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 100 {
		t.Errorf("expected balance 100 (in account currency USD), got %d", updated.Balance)
	}
}

func TestAddExpenseBaseCurrency(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	account := mustCreateAccount(t, m, "GoPay", "ewallet", "IDR")

	result, err := m.AddExpense(CreateExpenseParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -50000 {
		t.Errorf("expected balance -50000, got %d", updated.Balance)
	}
}

func TestAddExpenseForeignCurrencyEUR(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	account := mustCreateAccount(t, m, "Revolut EUR", "checking", "EUR")

	result, err := m.AddExpense(CreateExpenseParams{
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

	updated, err := m.q.GetAccountByID(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != -50 {
		t.Errorf("expected balance -50 (in EUR), got %d", updated.Balance)
	}
}

func TestAddExpenseMissingRate(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	_ = mustCreateAccount(t, m, "KRW Account", "checking", "KRW")

	_, err := m.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Korean food",
		Category:    "Restaurant",
		Account:     "KRW Account",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error for missing rate")
	}

	var rnf *shared.RateNotFoundError
	if !errors.As(err, &rnf) {
		t.Errorf("expected RateNotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "wallet rate add KRW") {
		t.Errorf("expected actionable error, got: %v", err)
	}

	transactions, err := m.ListTransactions(ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(transactions.Transactions) > 0 {
		t.Errorf("expected no transaction to be persisted, got %d", len(transactions.Transactions))
	}
}

func TestAddIncomeMissingRate(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	_ = mustCreateAccount(t, m, "JPY Account", "checking", "JPY")

	_, err := m.AddIncome(CreateIncomeParams{
		Amount:      5000,
		Description: "Japanese income",
		Category:    "Salary",
		Account:     "JPY Account",
		Date:        "2026-07-01",
	})
	if err == nil {
		t.Fatal("expected error for missing rate")
	}

	transactions, err := m.ListTransactions(ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(transactions.Transactions) > 0 {
		t.Errorf("expected no transaction to be persisted, got %d", len(transactions.Transactions))
	}
}

func TestListTransactionsWithOnlyBaseCurrency(t *testing.T) {
	m := setupManagerWithMultiCurrency(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")

	_, err := m.AddExpense(CreateExpenseParams{
		Amount:      50000,
		Description: "Coffee",
		Category:    "Coffee & Snacks",
		Account:     "BCA",
		Date:        "2026-07-01",
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	result, err := m.ListTransactions(ListTransactionsParams{Limit: 10})
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

func TestStringToInterface(t *testing.T) {
	if StringToInterface("") != nil {
		t.Error("expected nil for empty string")
	}
	if StringToInterface("hello") != "hello" {
		t.Error("expected string for non-empty string")
	}
}

func TestSumTransactionAmounts(t *testing.T) {
	if total := SumTransactionAmounts(nil); total != 0 {
		t.Errorf("expected 0 for nil, got %d", total)
	}
	if total := SumTransactionAmounts([]*gen.Transaction{}); total != 0 {
		t.Errorf("expected 0 for empty, got %d", total)
	}

	txns := []*gen.Transaction{
		{Amount: 100},
		{Amount: 200},
		{Amount: -50},
	}
	if total := SumTransactionAmounts(txns); total != 250 {
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

func TestResolveBaseFieldsMissingRateConfig(t *testing.T) {
	origLoad := shared.LoadRates
	shared.LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, errors.New("rate configuration not found")
	}
	defer func() { shared.LoadRates = origLoad }()

	m := NewManager(gen.New(testdb.Open(t)))
	_, _, err := m.ResolveBaseFields("USD", 100)
	if err == nil {
		t.Fatal("expected error for missing rate config in resolveBaseFields")
	}
}

type createFailQuerier struct {
	gen.Querier
}

func (c createFailQuerier) CreateTransaction(ctx context.Context, arg gen.CreateTransactionParams) (*gen.Transaction, error) {
	return nil, fmt.Errorf("mock create failure")
}

type updateFailQuerier struct {
	gen.Querier
}

func (u updateFailQuerier) UpdateTransaction(ctx context.Context, arg gen.UpdateTransactionParams) (*gen.Transaction, error) {
	return nil, fmt.Errorf("mock update failure")
}

type archiveFailQuerier struct {
	gen.Querier
}

func (a archiveFailQuerier) ArchiveTransaction(ctx context.Context, id int64) error {
	return fmt.Errorf("mock archive failure")
}

type balanceFailQuerier struct {
	gen.Querier
}

func (b balanceFailQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	return nil, fmt.Errorf("mock balance failure")
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

type balanceTypeQuerier struct {
	gen.Querier
}

func (b balanceTypeQuerier) GetAccountBalance(ctx context.Context, accountID int64) (interface{}, error) {
	return "not-an-int64", nil
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

type listTagsFailQuerier struct {
	gen.Querier
}

func (l listTagsFailQuerier) ListTransactionTags(ctx context.Context, transactionID int64) ([]*gen.Tag, error) {
	return nil, fmt.Errorf("mock list tags failure")
}

func TestAddExpense_DBError(t *testing.T) {
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)

	db := testdb.Open(t)
	q := gen.New(db)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_ = db.Close()

	failMgr := &Manager{q: q}

	_, err := failMgr.AddExpense(CreateExpenseParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)

	db := testdb.Open(t)
	q := gen.New(db)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_ = db.Close()

	failMgr := &Manager{q: q}

	_, err := failMgr.AddIncome(CreateIncomeParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)

	db := testdb.Open(t)
	q := gen.New(db)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	_ = db.Close()

	failMgr := &Manager{q: q}

	_, err := failMgr.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestListTransactions_DBError(t *testing.T) {
	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()

	failMgr := &Manager{q: q}

	_, err := failMgr.ListTransactions(ListTransactionsParams{
		Limit: 10,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestListTransactionsByTag_DBError(t *testing.T) {
	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()

	failMgr := &Manager{q: q}

	_, err := failMgr.ListTransactions(ListTransactionsParams{
		TagName: "food",
		Limit:   10,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestEditTransaction_DBError(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()
	failMgr := &Manager{q: q}

	_, err := failMgr.EditTransaction(txn.Transaction.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestRemoveTransaction_DBError(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	txn, _ := m.AddExpense(CreateExpenseParams{
		Amount:      35000,
		Description: "Lunch",
		Category:    "Restaurant",
		Account:     "BCA",
		Date:        "2026-07-01",
	})

	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()
	failMgr := &Manager{q: q}

	err := failMgr.RemoveTransaction(txn.Transaction.ID)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestAdjustBalance_DBError(t *testing.T) {
	m := setupManager(t)

	_ = mustCreateAccount(t, m, "BCA", "checking", "IDR")
	_, _ = m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})

	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()
	failMgr := &Manager{q: q}

	_, err := failMgr.AdjustBalance(AdjustBalanceParams{
		Account: "BCA",
		Target:  2000000,
	})
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestGetTransactionByID_DBError(t *testing.T) {
	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()
	failMgr := &Manager{q: q}

	_, err := failMgr.GetTransactionByID(1)
	if err == nil {
		t.Fatal("expected error with closed DB")
	}
}

func TestRecalculateBalanceDBError(t *testing.T) {
	db := testdb.Open(t)
	q := gen.New(db)
	_ = db.Close()
	failMgr := &Manager{q: q}

	err := failMgr.RecalcBalance(1)
	if err == nil {
		t.Fatal("expected error from recalculateBalance with closed DB")
	}
}

func TestAddExpenseCreateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: createFailQuerier{q}}

	_, err := m.AddExpense(CreateExpenseParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: createFailQuerier{q}}

	_, err := m.AddIncome(CreateIncomeParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	m := &Manager{q: createFailQuerier{q}}

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestEditTransactionUpdateFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: updateFailQuerier{q}}

	_, err := m.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestRemoveTransactionArchiveFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: archiveFailQuerier{q}}

	err := m.RemoveTransaction(txn.ID)
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestAdjustBalanceBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: balanceFailQuerier{q}}

	_, err := m.AdjustBalance(AdjustBalanceParams{
		Account: "BCA",
		Target:  1000000,
	})
	if err == nil {
		t.Fatal("expected error from mock")
	}
}

func TestAddExpenseTagAddFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "lunch")
	m := &Manager{q: tagAddFailQuerier{q}}

	_, err := m.AddExpense(CreateExpenseParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "salary-tag")
	m := &Manager{q: tagAddFailQuerier{q}}

	_, err := m.AddIncome(CreateIncomeParams{
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

func TestRecalculateBalanceTypeError(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: balanceTypeQuerier{q}}

	err := m.RecalcBalance(txn.AccountID)
	if err == nil {
		t.Fatal("expected type error from recalculateBalance")
	}
	if !strings.Contains(err.Error(), "unexpected balance type") {
		t.Errorf("expected 'unexpected balance type' error, got: %v", err)
	}
}

func TestAddExpenseBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: balanceTypeQuerier{q}}

	_, err := m.AddExpense(CreateExpenseParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: balanceTypeQuerier{q}}

	_, err := m.AddIncome(CreateIncomeParams{
		Amount:      1000000,
		Description: "Salary",
		Category:    "Salary",
		Account:     "BCA",
	})
	if err == nil {
		t.Fatal("expected balance recalc error")
	}
}

func TestAddTransferRecalcBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	m := &Manager{q: &transferBalanceFailQuerier{Querier: q}}

	_, err := m.AddTransfer(CreateTransferParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: balanceFailQuerier{q}}

	_, err := m.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected balance recalc error during edit")
	}
}

func TestRemoveTransactionRecalcBalanceFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: balanceFailQuerier{q}}

	err := m.RemoveTransaction(txn.ID)
	if err == nil {
		t.Fatal("expected balance recalc error during remove")
	}
}

func TestAdjustBalanceCreateTransactionFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: createFailQuerier{q}}

	_, err := m.AdjustBalance(AdjustBalanceParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "test-tag")
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: tagAddFailQuerier{q}}

	_, err := m.EditTransaction(txn.ID, EditTransactionParams{
		AddTagNames: []string{"test-tag"},
	})
	if err == nil {
		t.Fatal("expected tag add failure during edit")
	}
}

func TestEditTransactionTagRemoveFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateTag(context.Background(), "test-tag")
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	_ = q.AddTransactionTag(context.Background(), gen.AddTransactionTagParams{TransactionID: txn.ID, TagID: 1})

	m := &Manager{q: tagRemoveFailQuerier{q}}

	_, err := m.EditTransaction(txn.ID, EditTransactionParams{
		RemoveTagNames: []string{"test-tag"},
	})
	if err == nil {
		t.Fatal("expected tag remove failure during edit")
	}
}

func TestEditTransactionListTagsFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	txn, _ := q.CreateTransaction(context.Background(), gen.CreateTransactionParams{
		AccountID: 1, Type: "expense", Amount: 35000, Currency: "IDR", Date: "2026-07-01",
	})
	m := &Manager{q: listTagsFailQuerier{q}}

	_, err := m.EditTransaction(txn.ID, EditTransactionParams{})
	if err == nil {
		t.Fatal("expected list tags failure during edit")
	}
}

func TestAddTransferDestinationRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	m := &Manager{q: &transferDestFailQuerier{Querier: q}}

	_, err := m.AddTransfer(CreateTransferParams{
		Amount:      100000,
		FromAccount: "BCA",
		ToAccount:   "GoPay",
	})
	if err == nil {
		t.Fatal("expected destination recalc failure during transfer")
	}
}

func TestAdjustBalanceRecalcFailure(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: &adjustRecalcFailQuerier{Querier: q}}

	_, err := m.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1000000,
		Description: "Adjustment",
	})
	if err == nil {
		t.Fatal("expected adjust recalc failure")
	}
}

func TestAddTransferGetBalanceError(t *testing.T) {
	dbase := testdb.Open(t)
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "GoPay", Type: "ewallet", Currency: "IDR"})
	m := &Manager{q: transferFirstBalanceFailQuerier{q}}

	_, err := m.AddTransfer(CreateTransferParams{
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
	shared.SetTestRateConfig(shared.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})
	t.Cleanup(shared.ResetTestRateConfig)
	q := gen.New(dbase)
	_, _ = q.CreateAccount(context.Background(), gen.CreateAccountParams{Name: "BCA", Type: "checking", Currency: "IDR"})
	m := &Manager{q: &adjustFinalFailQuerier{Querier: q}}

	_, err := m.AdjustBalance(AdjustBalanceParams{
		Account:     "BCA",
		Target:      1000000,
		Description: "Adjustment",
	})
	if err == nil {
		t.Fatal("expected final balance read error")
	}
}
