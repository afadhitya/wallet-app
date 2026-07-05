package cli

import (
	"strings"
	"testing"
)

func setupAccountTest(t *testing.T) func() {
	cleanup := setupTestService()
	svc, _, _ := getServiceOverride(nil)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")
	return cleanup
}

func TestCLIAccountAdd(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "add", "Mandiri")
	if err != nil {
		t.Fatalf("account add: %v", err)
	}
	if !strings.Contains(stdout, "Account 'Mandiri' created") {
		t.Errorf("expected creation message, got: %s", stdout)
	}
}

func TestCLIAccountAddJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("--json", "account", "add", "Mandiri")
	if err != nil {
		t.Fatalf("account add --json: %v", err)
	}
	result := extractJSONData(t, stdout)
	if result["name"] != "Mandiri" {
		t.Errorf("expected name 'Mandiri', got %v", result["name"])
	}
	if result["type"] != "checking" {
		t.Errorf("expected type 'checking', got %v", result["type"])
	}
}

func TestCLIAccountAddWithType(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "add", "Savings", "--type", "savings", "--currency", "USD")
	if err != nil {
		t.Fatalf("account add with type: %v", err)
	}
	if !strings.Contains(stdout, "Account 'Savings' created") {
		t.Errorf("expected creation message, got: %s", stdout)
	}
}

func TestCLIAccountAddDuplicate(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "add", "BCA")
	if err == nil {
		t.Fatal("expected error for duplicate account name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected duplicate error, got: %v", err)
	}
}

func TestCLIAccountAddDuplicateJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, stderr, err := runTestCmd("--json", "account", "add", "BCA")
	if err == nil {
		t.Fatal("expected error for duplicate account name with --json")
	}
	if !strings.Contains(stderr, "success") || !strings.Contains(stderr, "false") {
		t.Errorf("expected JSON error response, got: %s", stderr)
	}
}

func TestCLIAccountAddEmptyName(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "add", "")
	if err == nil {
		t.Fatal("expected error for empty account name")
	}
}

func TestCLIAccountList(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "list")
	if err != nil {
		t.Fatalf("account list: %v", err)
	}
	if !strings.Contains(stdout, "BCA") {
		t.Errorf("expected BCA in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "GoPay") {
		t.Errorf("expected GoPay in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Total") {
		t.Errorf("expected Total row in output, got: %s", stdout)
	}
}

func TestCLIAccountListJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("--json", "account", "list")
	if err != nil {
		t.Fatalf("account list --json: %v", err)
	}
	arr := extractJSONArray(t, stdout)
	if len(arr) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(arr))
	}
}

func TestCLIAccountListAll(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	svc, _, _ := getServiceOverride(nil)
	_ = svc.ArchiveAccount(1)

	stdout, _, err := runTestCmd("account", "list", "--all")
	if err != nil {
		t.Fatalf("account list --all: %v", err)
	}
	if !strings.Contains(stdout, "BCA") {
		t.Errorf("expected archived BCA in output with --all, got: %s", stdout)
	}
	if !strings.Contains(stdout, "archived") {
		t.Errorf("expected 'archived' status in output, got: %s", stdout)
	}
}

func TestCLIAccountListAllJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	svc, _, _ := getServiceOverride(nil)
	_ = svc.ArchiveAccount(1)

	stdout, _, err := runTestCmd("--json", "account", "list", "--all")
	if err != nil {
		t.Fatalf("account list --all --json: %v", err)
	}
	arr := extractJSONArray(t, stdout)
	if len(arr) != 2 {
		t.Errorf("expected 2 accounts (including archived), got %d", len(arr))
	}
}

func TestCLIAccountListEmpty(t *testing.T) {
	cleanup := setupTestService()
	defer cleanup()

	stdout, _, err := runTestCmd("account", "list")
	if err != nil {
		t.Fatalf("account list: %v", err)
	}
	if !strings.Contains(stdout, "No accounts found") {
		t.Errorf("expected 'No accounts found', got: %s", stdout)
	}
}

func TestCLIAccountEdit(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "edit", "1", "--name", "BCA Main")
	if err != nil {
		t.Fatalf("account edit: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 updated") {
		t.Errorf("expected edit success message, got: %s", stdout)
	}
}

func TestCLIAccountEditJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("--json", "account", "edit", "1", "--name", "BCA Main")
	if err != nil {
		t.Fatalf("account edit --json: %v", err)
	}
	result := extractJSONData(t, stdout)
	if result["name"] != "BCA Main" {
		t.Errorf("expected name 'BCA Main', got %v", result["name"])
	}
}

func TestCLIAccountEditType(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "edit", "1", "--type", "savings")
	if err != nil {
		t.Fatalf("account edit --type: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 updated") {
		t.Errorf("expected edit success message, got: %s", stdout)
	}

	svc, _, _ := getServiceOverride(nil)
	account, _ := svc.GetAccountByID(1)
	if account.Type != "savings" {
		t.Errorf("expected type 'savings', got '%s'", account.Type)
	}
}

func TestCLIAccountEditSortOrder(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "edit", "1", "--sort-order", "5")
	if err != nil {
		t.Fatalf("account edit --sort-order: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 updated") {
		t.Errorf("expected edit success message, got: %s", stdout)
	}
}

func TestCLIAccountEditNotFound(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "edit", "99", "--name", "Ghost")
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCLIAccountEditEmptyName(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "edit", "1", "--name", "")
	if err == nil {
		t.Fatal("expected error for empty name on edit")
	}
}

func TestCLIAccountEditNoChanges(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "edit", "1")
	if err == nil {
		t.Fatal("expected error for edit with no flags")
	}
	if !strings.Contains(err.Error(), "at least one field") {
		t.Errorf("expected 'at least one field' error, got: %v", err)
	}
}

func TestCLIAccountEditInvalidID(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "edit", "abc", "--name", "Test")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(err.Error(), "invalid account ID") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestCLIAccountArchive(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "archive", "1", "--force")
	if err != nil {
		t.Fatalf("account archive: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 archived") {
		t.Errorf("expected archive success message, got: %s", stdout)
	}
}

func TestCLIAccountArchiveJSON(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("--json", "account", "archive", "1", "--force")
	if err != nil {
		t.Fatalf("account archive --json: %v", err)
	}
	result := extractJSONData(t, stdout)
	if result["status"] != "archived" {
		t.Errorf("expected status 'archived', got %v", result["status"])
	}
}

func TestCLIAccountArchiveNotFound(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "archive", "99", "--force")
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCLIAccountArchiveInvalidID(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	_, _, err := runTestCmd("account", "archive", "abc", "--force")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(err.Error(), "invalid account ID") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestCLIAccountArchiveBalanceWarning(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	svc, _, _ := getServiceOverride(nil)
	_ = svc.UpdateAccountBalance(1, 50000)

	stdout, _, err := runTestCmd("account", "archive", "1", "--force")
	if err != nil {
		t.Fatalf("account archive with balance: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 archived") {
		t.Errorf("expected archive success message, got: %s", stdout)
	}
}

func TestCLIAccountArchiveConfirmation(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	oldStdin := osStdin
	osStdin = strings.NewReader("yes\n")
	defer func() { osStdin = oldStdin }()

	stdout, _, err := runTestCmd("account", "archive", "1")
	if err != nil {
		t.Fatalf("account archive with confirmation: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 archived") {
		t.Errorf("expected archive success message, got: %s", stdout)
	}
}

func TestCLIAccountArchiveConfirmationDecline(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	oldStdin := osStdin
	osStdin = strings.NewReader("no\n")
	defer func() { osStdin = oldStdin }()

	stdout, _, err := runTestCmd("account", "archive", "1")
	if err != nil {
		t.Fatalf("account archive with decline: %v", err)
	}
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("expected 'Cancelled' message, got: %s", stdout)
	}
}

func TestCLIAccountArchiveBalanceWarningWithConfirmation(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	svc, _, _ := getServiceOverride(nil)
	_ = svc.UpdateAccountBalance(1, 50000)

	oldStdin := osStdin
	osStdin = strings.NewReader("yes\n")
	defer func() { osStdin = oldStdin }()

	stdout, _, err := runTestCmd("account", "archive", "1")
	if err != nil {
		t.Fatalf("account archive with balance warning: %v", err)
	}
	if !strings.Contains(stdout, "non-zero balance") {
		t.Errorf("expected balance warning, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Account 1 archived") {
		t.Errorf("expected archive success message, got: %s", stdout)
	}
}

func TestCLIAccountHelp(t *testing.T) {
	cleanup := setupAccountTest(t)
	defer cleanup()

	stdout, _, err := runTestCmd("account", "--help")
	if err != nil {
		t.Fatalf("account --help: %v", err)
	}
	for _, sub := range []string{"add", "list", "edit", "archive"} {
		if !strings.Contains(stdout, sub) {
			t.Errorf("expected help output to contain '%s'", sub)
		}
	}
}
