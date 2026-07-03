package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/service"
)

type testCLI struct {
	t       *testing.T
	svc     *service.Service
	cleanup func()
}

func newTestCLI(t *testing.T) *testCLI {
	t.Helper()
	cleanup := setupTestService()
	t.Cleanup(cleanup)

	svc, _, _ := getServiceOverride(nil)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")

	return &testCLI{t: t, svc: svc, cleanup: cleanup}
}

func (c *testCLI) run(args ...string) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCLIInit(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !strings.Contains(stdout, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", stdout)
	}
}

func TestCLIInitJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "init")
	if err != nil {
		t.Fatalf("init --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "initialized" {
		t.Errorf("expected status 'initialized', got %v", result["status"])
	}
}

func TestCLICategoryList(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("category", "list")
	if err != nil {
		t.Fatalf("category list: %v", err)
	}
	if !strings.Contains(stdout, "Food & Dining") {
		t.Errorf("expected 'Food & Dining' in output")
	}
}

func TestCLICategoryListJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "category", "list")
	if err != nil {
		t.Fatalf("category list --json: %v", err)
	}

	var categories []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &categories); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if len(categories) < 10 {
		t.Errorf("expected at least 10 categories, got %d", len(categories))
	}
}

func TestCLICategoryAdd(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("category", "add", "Kopi", "--icon", "coffee")
	if err != nil {
		t.Fatalf("category add: %v", err)
	}
	if !strings.Contains(stdout, "Kopi") {
		t.Errorf("expected 'Kopi' in output: %s", stdout)
	}
}

func TestCLICategoryEdit(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("category", "add", "TestCat")

	stdout, _, err := cli.run("category", "edit", "1", "--name", "Modified")
	if err != nil {
		t.Fatalf("category edit: %v", err)
	}
	if !strings.Contains(stdout, "updated") {
		t.Errorf("expected 'updated' in output: %s", stdout)
	}
}

func TestCLICategoryRm(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("category", "add", "TempCat")

	stdout, _, err := cli.run("category", "rm", "1")
	if err != nil {
		t.Fatalf("category rm: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output: %s", stdout)
	}
}

func TestCLITagList(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("tag", "list")
	if err != nil {
		t.Fatalf("tag list: %v", err)
	}
	if !strings.Contains(stdout, "No tags") {
		t.Errorf("expected 'No tags' for empty list")
	}
}

func TestCLITagAdd(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("tag", "add", "japan-2026")
	if err != nil {
		t.Fatalf("tag add: %v", err)
	}
	if !strings.Contains(stdout, "japan-2026") {
		t.Errorf("expected 'japan-2026' in output: %s", stdout)
	}
}

func TestCLITagRm(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "temp")

	stdout, _, err := cli.run("tag", "rm", "temp")
	if err != nil {
		t.Fatalf("tag rm: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output: %s", stdout)
	}
}

func TestCLIListNoTransactions(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "No transactions") {
		t.Errorf("expected 'No transactions' for empty list")
	}
}

func TestCLIListJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "list")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["count"].(float64) != 0 {
		t.Errorf("expected count 0, got %v", result["count"])
	}
}

func TestCLIAddExpense(t *testing.T) {
	cli := newTestCLI(t)
	_, _, err := cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}

	stdout, _, err := cli.run("list")
	if err != nil {
		t.Fatalf("list after expense: %v", err)
	}
	if !strings.Contains(stdout, "Lunch") {
		t.Errorf("expected 'Lunch' in list output: %s", stdout)
	}
}

func TestCLIAddIncome(t *testing.T) {
	cli := newTestCLI(t)
	_, _, err := cli.run("add", "income", "5000000", "Gaji", "-c", "Salary", "-a", "BCA")
	if err != nil {
		t.Fatalf("add income: %v", err)
	}

	stdout, _, err := cli.run("list")
	if err != nil {
		t.Fatalf("list after income: %v", err)
	}
	if !strings.Contains(stdout, "Gaji") {
		t.Errorf("expected 'Gaji' in list output: %s", stdout)
	}
}

func TestCLIAddTransfer(t *testing.T) {
	cli := newTestCLI(t)
	_, _, err := cli.run("add", "transfer", "200000", "--from", "BCA", "--to", "GoPay")
	if err != nil {
		t.Fatalf("add transfer: %v", err)
	}

	stdout, _, err := cli.run("list", "--type", "transfer")
	if err != nil {
		t.Fatalf("list after transfer: %v", err)
	}
	if !strings.Contains(stdout, "transfer") {
		t.Errorf("expected 'transfer' in list output: %s", stdout)
	}
}

func TestCLIEditTransaction(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("edit", "1", "--amount", "40000")
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !strings.Contains(stdout, "updated") {
		t.Errorf("expected 'updated' in output: %s", stdout)
	}
}

func TestCLIRemoveTransaction(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("rm", "1", "--force")
	if err != nil {
		t.Fatalf("rm --force: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output: %s", stdout)
	}
}

func TestCLIAdjust(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "income", "1000000", "Initial", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("adjust", "BCA", "1500000", "Correction")
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if !strings.Contains(stdout, "Difference") {
		t.Errorf("expected 'Difference' in output: %s", stdout)
	}
}

func TestCLIMissingAccount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "Ghost")
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "account") {
		t.Errorf("expected account not found error, got: %s", stderr)
	}
}

func TestCLIMissingCategory(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "35000", "Lunch", "-c", "Ghost", "-a", "BCA")
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "category") {
		t.Errorf("expected category not found error, got: %s", stderr)
	}
}

func TestCLIMissingTag(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA", "-t", "ghost")
	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "tag") {
		t.Errorf("expected tag not found error, got: %s", stderr)
	}
}

func TestCLISameAccountTransfer(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "transfer", "100000", "--from", "BCA", "--to", "BCA")
	if !strings.Contains(stderr, "different") {
		t.Errorf("expected 'different' error, got: %s", stderr)
	}
}
