package cli

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func replaceStdin(t *testing.T, input string) func() {
	t.Helper()
	oldStdin := osStdin
	osStdin = strings.NewReader(input)
	return func() {
		osStdin = oldStdin
	}
}

func replaceStdinEmpty(t *testing.T) func() {
	t.Helper()
	oldStdin := osStdin
	osStdin = &eofReader{}
	return func() {
		osStdin = oldStdin
	}
}

type eofReader struct{}

func (e *eofReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func TestCLIAddExpense_JSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	if err != nil {
		t.Fatalf("add expense --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["type"] != "expense" {
		t.Errorf("expected type 'expense', got %v", result["type"])
	}
}

func TestCLIAddExpense_JSON_WithTag(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "japan-2026")
	stdout, _, err := cli.run("--json", "add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA", "-t", "japan-2026")
	if err != nil {
		t.Fatalf("add expense --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v", jerr)
	}
	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected tags array, got %T", result["tags"])
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
	if tags[0] != "japan-2026" {
		t.Errorf("expected tag 'japan-2026', got %v", tags[0])
	}
}

func TestCLIAddExpense_WithTag(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "japan-2026")
	stdout, _, err := cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA", "-t", "japan-2026")
	if err != nil {
		t.Fatalf("add expense with tag: %v", err)
	}
	if !strings.Contains(stdout, "Tags:") {
		t.Errorf("expected 'Tags:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "japan-2026") {
		t.Errorf("expected 'japan-2026' in output, got: %s", stdout)
	}
}

func TestCLIAddIncome_JSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "add", "income", "5000000", "Gaji", "-c", "Salary", "-a", "BCA")
	if err != nil {
		t.Fatalf("add income --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v", jerr)
	}
	if result["type"] != "income" {
		t.Errorf("expected type 'income', got %v", result["type"])
	}
}

func TestCLIAddTransfer_JSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "add", "transfer", "200000", "--from", "BCA", "--to", "GoPay")
	if err != nil {
		t.Fatalf("add transfer --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["type"] != "transfer" {
		t.Errorf("expected type 'transfer', got %v", result["type"])
	}
}

func TestCLIAddTransfer_NonJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("add", "transfer", "200000", "--from", "BCA", "--to", "GoPay")
	if err != nil {
		t.Fatalf("add transfer: %v", err)
	}
	if !strings.Contains(stdout, "Transfer recorded") {
		t.Errorf("expected 'Transfer recorded' in output, got: %s", stdout)
	}
}

func TestCLIEdit_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("--json", "edit", "1", "--amount", "40000")
	if err != nil {
		t.Fatalf("edit --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["amount"].(float64) != 40000 {
		t.Errorf("expected amount 40000, got %v", result["amount"])
	}
}

func TestCLIRm_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("--json", "rm", "1", "--force")
	if err != nil {
		t.Fatalf("rm --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLIAdjust_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "income", "1000000", "Initial", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("--json", "adjust", "BCA", "1500000", "Correction")
	if err != nil {
		t.Fatalf("adjust --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["account"] != "BCA" {
		t.Errorf("expected account 'BCA', got %v", result["account"])
	}
	if result["description"] != "Correction" {
		t.Errorf("expected description 'Correction', got %v", result["description"])
	}
}

func TestCLICategoryAdd_JSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "category", "add", "Kopi", "--icon", "coffee")
	if err != nil {
		t.Fatalf("category add --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["name"] != "Kopi" {
		t.Errorf("expected name 'Kopi', got %v", result["name"])
	}
}

func TestCLICategoryEdit_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("category", "add", "TestCat")

	stdout, _, err := cli.run("--json", "category", "edit", "1", "-n", "Modified")
	if err != nil {
		t.Fatalf("category edit --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["name"] != "Modified" {
		t.Errorf("expected name 'Modified', got %v", result["name"])
	}
}

func TestCLICategoryRm_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("category", "add", "TempCat")

	stdout, _, err := cli.run("--json", "category", "rm", "1")
	if err != nil {
		t.Fatalf("category rm --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLITagList_WithTags(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "vip")
	_, _, _ = cli.run("tag", "add", "recurring")

	stdout, _, err := cli.run("tag", "list")
	if err != nil {
		t.Fatalf("tag list: %v", err)
	}
	if !strings.Contains(stdout, "vip") {
		t.Errorf("expected 'vip' in output: %s", stdout)
	}
	if !strings.Contains(stdout, "recurring") {
		t.Errorf("expected 'recurring' in output: %s", stdout)
	}
}

func TestCLITagList_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "vip")

	stdout, _, err := cli.run("--json", "tag", "list")
	if err != nil {
		t.Fatalf("tag list --json: %v", err)
	}

	var tags []map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &tags); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestCLITagList_JSON_Empty(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "tag", "list")
	if err != nil {
		t.Fatalf("tag list --json: %v", err)
	}

	var tags []map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &tags); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestCLITagAdd_JSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "tag", "add", "japan-2026")
	if err != nil {
		t.Fatalf("tag add --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["name"] != "japan-2026" {
		t.Errorf("expected name 'japan-2026', got %v", result["name"])
	}
}

func TestCLITagRm_JSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "temp")

	stdout, _, err := cli.run("--json", "tag", "rm", "temp")
	if err != nil {
		t.Fatalf("tag rm --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v, output: %s", jerr, stdout)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLIAddExpense_InvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "abc", "Lunch", "-c", "Restaurant", "-a", "BCA")
	if !strings.Contains(stderr, "invalid amount") {
		t.Errorf("expected 'invalid amount' error, got: %s", stderr)
	}
}

func TestCLIAddIncome_InvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "income", "xyz", "Gaji", "-c", "Salary", "-a", "BCA")
	if !strings.Contains(stderr, "invalid amount") {
		t.Errorf("expected 'invalid amount' error, got: %s", stderr)
	}
}

func TestCLIAddTransfer_InvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "transfer", "bad_amount", "--from", "BCA", "--to", "GoPay")
	if !strings.Contains(stderr, "invalid amount") {
		t.Errorf("expected 'invalid amount' error, got: %s", stderr)
	}
}

func TestCLIEdit_InvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("edit", "abc")
	if !strings.Contains(stderr, "invalid transaction ID") {
		t.Errorf("expected 'invalid transaction ID' error, got: %s", stderr)
	}
}

func TestCLIEdit_InvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	_, stderr, _ := cli.run("edit", "1", "--amount", "xyz")
	if !strings.Contains(stderr, "invalid amount") {
		t.Errorf("expected 'invalid amount' error, got: %s", stderr)
	}
}

func TestCLIEdit_NonExistent(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("edit", "999", "--amount", "5000")
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "notFound") && !strings.Contains(stderr, "exists") {
		t.Errorf("expected not-found error, got: %s", stderr)
	}
}

func TestCLIRm_InvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("rm", "abc", "--force")
	if !strings.Contains(stderr, "invalid transaction ID") {
		t.Errorf("expected 'invalid transaction ID' error, got: %s", stderr)
	}
}

func TestCLIRm_NonExistent(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("rm", "999", "--force")
	if err == nil {
		t.Fatal("expected error for non-existent transaction")
	}
	if stderr == "" && err != nil {
		t.Logf("rm non-existent error: %v", err)
	}
}

func TestCLIRm_ConfirmYes(t *testing.T) {
	cleanup := replaceStdin(t, "yes\n")
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("rm", "1")
	if err != nil {
		t.Fatalf("rm with confirmation 'yes': %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", stdout)
	}
}

func TestCLIRm_ConfirmY(t *testing.T) {
	cleanup := replaceStdin(t, "y\n")
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("rm", "1")
	if err != nil {
		t.Fatalf("rm with confirmation 'y': %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", stdout)
	}
}

func TestCLIRm_ConfirmNo(t *testing.T) {
	cleanup := replaceStdin(t, "no\n")
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("rm", "1")
	if err != nil {
		t.Fatalf("rm with confirmation 'no': %v", err)
	}
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("expected 'Cancelled' in output, got: %s", stdout)
	}
}

func TestCLIRm_ConfirmCancel(t *testing.T) {
	cleanup := replaceStdin(t, "whatever\n")
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("rm", "1")
	if err != nil {
		t.Fatalf("rm with confirmation 'whatever': %v", err)
	}
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("expected 'Cancelled' in output, got: %s", stdout)
	}
}

func TestCLIRm_ConfirmStdinError(t *testing.T) {
	cleanup := replaceStdinEmpty(t)
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	_, stderr, err := cli.run("rm", "1")
	if err == nil {
		t.Fatal("expected error when stdin is closed")
	}
	if !strings.Contains(stderr, "failed to read confirmation") {
		t.Errorf("expected 'failed to read confirmation' error, got: %s", stderr)
	}
}

func TestCLIAdjust_InvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("adjust", "BCA", "xyz", "Bad amount")
	if !strings.Contains(stderr, "invalid amount") {
		t.Errorf("expected 'invalid amount' error, got: %s", stderr)
	}
}

func TestCLIAdjust_NonExistentAccount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("adjust", "GhostAccount", "100000", "Correction")
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "account") && !strings.Contains(stderr, "exists") {
		t.Errorf("expected not-found error, got: %s", stderr)
	}
}

func TestCLICategoryEdit_InvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("category", "edit", "abc", "-n", "NewName")
	if !strings.Contains(stderr, "invalid category ID") {
		t.Errorf("expected 'invalid category ID' error, got: %s", stderr)
	}
}

func TestCLICategoryEdit_NonExistent(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("category", "edit", "999", "-n", "GhostCat")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "exists") && !strings.Contains(stderr, "category") {
		t.Errorf("expected not-found error, got: %s", stderr)
	}
}

func TestCLICategoryRm_InvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("category", "rm", "xyz")
	if !strings.Contains(stderr, "invalid category ID") {
		t.Errorf("expected 'invalid category ID' error, got: %s", stderr)
	}
}

func TestCLICategoryRm_NonExistent(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("category", "rm", "999")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "exists") && !strings.Contains(stderr, "category") {
		t.Errorf("expected not-found error, got: %s", stderr)
	}
}

func TestCLITagRm_NonExistent(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("tag", "rm", "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "exists") {
		t.Errorf("expected not-found error, got: %s", stderr)
	}
}

func TestCLIList_WithFilters(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	_, _, _ = cli.run("add", "income", "5000000", "Gaji", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("list", "--type", "expense")
	if err != nil {
		t.Fatalf("list --type expense: %v", err)
	}
	if !strings.Contains(stdout, "Lunch") {
		t.Errorf("expected 'Lunch' in output: %s", stdout)
	}
	if strings.Contains(stdout, "Gaji") {
		t.Errorf("did not expect 'Gaji' when filtering by expense")
	}
}

func TestCLIList_WithAccountFilter(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("list", "--account", "BCA")
	if err != nil {
		t.Fatalf("list --account BCA: %v", err)
	}
	if !strings.Contains(stdout, "Lunch") {
		t.Errorf("expected 'Lunch' in output: %s", stdout)
	}
}

func TestCLIList_WithCategoryFilter(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("list", "--category", "Restaurant")
	if err != nil {
		t.Fatalf("list --category: %v", err)
	}
	if !strings.Contains(stdout, "Lunch") {
		t.Errorf("expected 'Lunch' in output: %s", stdout)
	}
}

func TestCLIList_WithDateRange(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("list", "--from", "2020-01-01", "--to", "2099-12-31")
	if err != nil {
		t.Fatalf("list with date range: %v", err)
	}
	if !strings.Contains(stdout, "Lunch") {
		t.Errorf("expected 'Lunch' in output: %s", stdout)
	}
}

func TestCLICategoryList_Empty(t *testing.T) {
	t.Skip("Categories are pre-seeded by migrations, so list is never empty")
}

func TestCLIInit_NonJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !strings.Contains(stdout, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Database:") {
		t.Errorf("expected 'Database:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Accounts:") {
		t.Errorf("expected 'Accounts:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Categories:") {
		t.Errorf("expected 'Categories:' in output, got: %s", stdout)
	}
}

func TestCLIAddExpense_TagNotFound(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA", "-t", "ghost")
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "tag") {
		t.Errorf("expected tag not found error, got: %s", stderr)
	}
}

func TestCLIEdit_WithTags(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "vip")
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("edit", "1", "--add-tag", "vip")
	if err != nil {
		t.Fatalf("edit --add-tag: %v", err)
	}
	if !strings.Contains(stdout, "updated") {
		t.Errorf("expected 'updated' in output: %s", stdout)
	}
}

func TestCLIAddIncome_NonExistentAccount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "income", "5000000", "Gaji", "-c", "Salary", "-a", "Ghost")
	if stderr == "" {
		t.Error("expected error for non-existent account in income")
	}
}

func TestCLICategoryAdd_DuplicateName(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("category", "add", "MyUniqueCat")

	_, stderr, err := cli.run("category", "add", "MyUniqueCat")
	if err == nil {
		t.Fatal("expected error for duplicate category name")
	}
	if !strings.Contains(stderr, "already exists") && !strings.Contains(stderr, "name already exists") {
		t.Errorf("expected duplicate name error, got: %s", stderr)
	}
}

func TestCLITagAdd_DuplicateName(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "duptag")

	_, stderr, err := cli.run("tag", "add", "duptag")
	if err == nil {
		t.Fatal("expected error for duplicate tag name")
	}
	if !strings.Contains(stderr, "already exists") && !strings.Contains(stderr, "name already exists") {
		t.Errorf("expected duplicate name error, got: %s", stderr)
	}
}

func TestCLIList_IncludesTransferForCoverage(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "transfer", "200000", "--from", "BCA", "--to", "GoPay")

	stdout, _, err := cli.run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "transfer") {
		t.Errorf("expected 'transfer' in output: %s", stdout)
	}
}

func TestCLIRm_NonExistentForce(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, err := cli.run("rm", "999", "--force")
	if err == nil {
		t.Fatal("expected error for non-existent transaction with force")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr)
	}
}

func TestCLIList_ErrorPath(t *testing.T) {
	oldOverride := getServiceOverride
	defer func() { getServiceOverride = oldOverride }()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	svc, _, _ := getServiceOverride(nil)
	_ = svc.DB().Close()

	_, stderr, _ := cli.run("list")
	if stderr == "" {
		t.Error("expected error output for list with closed DB")
	}
}

func TestCLIInit_JSONOutput(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "init")
	if err != nil {
		t.Fatalf("init --json: %v", err)
	}

	var result map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &result); jerr != nil {
		t.Fatalf("unmarshal JSON: %v", jerr)
	}
}

func TestCLITagList_EmptyJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "tag", "list")
	if err != nil {
		t.Fatalf("tag list --json: %v", err)
	}

	var tags []map[string]interface{}
	if jerr := json.Unmarshal([]byte(stdout), &tags); jerr != nil {
		t.Fatalf("unmarshal JSON: %v", jerr)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestCLITagList_NonEmpty(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "alpha")
	_, _, _ = cli.run("tag", "add", "beta")

	stdout, _, err := cli.run("tag", "list")
	if err != nil {
		t.Fatalf("tag list: %v", err)
	}
	if !strings.Contains(stdout, "alpha") || !strings.Contains(stdout, "beta") {
		t.Errorf("expected both tags in output: %s", stdout)
	}
}

func TestCLITagRm_ErrorPath(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("tag", "rm", "nonexistent")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr)
	}
}

func TestCLICategoryList_ClosedDB(t *testing.T) {
	oldOverride := getServiceOverride
	defer func() { getServiceOverride = oldOverride }()

	cli := newTestCLI(t)
	svc, _, _ := getServiceOverride(nil)
	_ = svc.DB().Close()

	_, stderr, _ := cli.run("category", "list")
	if stderr == "" {
		t.Error("expected error output for category list with closed DB")
	}
}

func TestCLIRm_StdinError(t *testing.T) {
	cleanup := replaceStdinEmpty(t)
	defer cleanup()

	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	_, stderr, err := cli.run("rm", "1")
	if err == nil {
		t.Fatal("expected error when stdin is closed during rm")
	}
	if !strings.Contains(stderr, "failed to read confirmation") {
		t.Errorf("expected 'failed to read confirmation', got: %s", stderr)
	}
}
