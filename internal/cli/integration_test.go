package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
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

func TestExpandHomePath(t *testing.T) {
	if result := expandHomePath("/absolute/path"); result != "/absolute/path" {
		t.Errorf("expected '/absolute/path', got '%s'", result)
	}
	if result := expandHomePath("simple/path"); result != "simple/path" {
		t.Errorf("expected 'simple/path', got '%s'", result)
	}
}

func TestFormatAmount(t *testing.T) {
	if result := formatAmount(0); result != "Rp 0" {
		t.Errorf("expected 'Rp 0', got '%s'", result)
	}
	if result := formatAmount(-1000); result != "-Rp 1.000" {
		t.Errorf("expected '-Rp 1.000', got '%s'", result)
	}
	if result := formatAmount(1000000); result != "Rp 1.000.000" {
		t.Errorf("expected 'Rp 1.000.000', got '%s'", result)
	}
}

func TestCLIRmJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("--json", "rm", "1", "--force")
	if err != nil {
		t.Fatalf("rm --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLITagListJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "test-tag")

	stdout, _, err := cli.run("--json", "tag", "list")
	if err != nil {
		t.Fatalf("tag list --json: %v", err)
	}

	var tags []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &tags); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestCLITagRmJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "temp")

	stdout, _, err := cli.run("--json", "tag", "rm", "temp")
	if err != nil {
		t.Fatalf("tag rm --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLITagRmNameNotFound(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("tag", "rm", "nonexistent")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr)
	}
}

func TestCLIAdjustJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "income", "1000000", "Initial", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("--json", "adjust", "BCA", "1500000", "Correction")
	if err != nil {
		t.Fatalf("adjust --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["difference"].(float64) != 500000 {
		t.Errorf("expected difference 500000, got %v", result["difference"])
	}
}

func TestCLIAddExpenseJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	if err != nil {
		t.Fatalf("add expense --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["type"] != "expense" {
		t.Errorf("expected type 'expense', got %v", result["type"])
	}
}

func TestCLIAddTransferJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "add", "transfer", "200000", "--from", "BCA", "--to", "GoPay")
	if err != nil {
		t.Fatalf("add transfer --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["type"] != "transfer" {
		t.Errorf("expected type 'transfer', got %v", result["type"])
	}
}

func TestCLIEditJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("--json", "edit", "1", "--amount", "40000")
	if err != nil {
		t.Fatalf("edit --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIEditInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("edit", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIEditInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	_, stderr, _ := cli.run("edit", "1", "--amount", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIRmInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("rm", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIAdjustInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "income", "1000000", "Initial", "-c", "Salary", "-a", "BCA")

	_, stderr, _ := cli.run("adjust", "BCA", "not-a-number", "Test")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLICategoryEditInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("category", "edit", "not-a-number", "--name", "Test")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLICategoryRmInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("category", "rm", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIAddExpenseInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "expense", "not-a-number", "Lunch", "-c", "Restaurant", "-a", "BCA")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIAddIncomeInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "income", "not-a-number", "Salary", "-c", "Salary", "-a", "BCA")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIAddTransferInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("add", "transfer", "not-a-number", "--from", "BCA", "--to", "GoPay")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIListTypeFilter(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	_, _, _ = cli.run("add", "income", "1000000", "Salary", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("list", "--type", "income")
	if err != nil {
		t.Fatalf("list --type: %v", err)
	}
	if strings.Contains(stdout, "Lunch") {
		t.Error("expected only income, but found Lunch")
	}
}

func TestCLIListDateRange(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "First", "-c", "Restaurant", "-a", "BCA", "-d", "2026-07-01")
	_, _, _ = cli.run("add", "expense", "15000", "Second", "-c", "Restaurant", "-a", "BCA", "-d", "2026-07-10")
	_, _, _ = cli.run("add", "expense", "25000", "Third", "-c", "Restaurant", "-a", "BCA", "-d", "2026-07-20")

	stdout, _, err := cli.run("list", "--from", "2026-07-01", "--to", "2026-07-10")
	if err != nil {
		t.Fatalf("list date range: %v", err)
	}
	if !strings.Contains(stdout, "First") {
		t.Error("expected 'First' in output")
	}
	if strings.Contains(stdout, "Third") {
		t.Error("expected no 'Third' in output")
	}
}

func TestCLIListAccountFilter(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("add", "expense", "35000", "BCA Expense", "-c", "Restaurant", "-a", "BCA")
	_, _, _ = cli.run("add", "expense", "15000", "GoPay Expense", "-c", "Restaurant", "-a", "GoPay")

	stdout, _, err := cli.run("list", "--account", "BCA")
	if err != nil {
		t.Fatalf("list --account: %v", err)
	}
	if strings.Contains(stdout, "GoPay Expense") {
		t.Error("expected only BCA, but found GoPay")
	}
}

func TestCLIListLimit(t *testing.T) {
	cli := newTestCLI(t)
	for i := 0; i < 5; i++ {
		_, _, _ = cli.run("add", "expense", fmt.Sprintf("%d", 10000+i), fmt.Sprintf("Expense %d", i), "-c", "Restaurant", "-a", "BCA", "-d", "2026-07-01")
	}

	stdout, _, err := cli.run("list", "--limit", "3")
	if err != nil {
		t.Fatalf("list --limit: %v", err)
	}
	if strings.Count(stdout, "Expense") != 3 {
		t.Errorf("expected 3 entries, got %d", strings.Count(stdout, "Expense"))
	}
}

func TestCLIBudgetSet(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("budget", "set", "Monthly Food", "2000000", "-c", "Restaurant", "--period", "monthly")
	if err != nil {
		t.Fatalf("budget set: %v", err)
	}
	if !strings.Contains(stdout, "Monthly Food") {
		t.Errorf("expected 'Monthly Food' in output: %s", stdout)
	}
	if !strings.Contains(stdout, "set") {
		t.Errorf("expected 'set' in output: %s", stdout)
	}
}

func TestCLIBudgetSetJSON(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "budget", "set", "Monthly Food", "2000000", "-c", "Restaurant", "--period", "monthly")
	if err != nil {
		t.Fatalf("budget set --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["name"] != "Monthly Food" {
		t.Errorf("expected name 'Monthly Food', got %v", result["name"])
	}
	if result["period"] != "monthly" {
		t.Errorf("expected period 'monthly', got %v", result["period"])
	}
}

func TestCLIBudgetSetInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "set", "Food", "not-a-number", "-c", "Restaurant")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIBudgetSetNoTargets(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "set", "Food", "1000000", "--period", "monthly")
	if !strings.Contains(stderr, "target") {
		t.Errorf("expected 'target' error, got: %s", stderr)
	}
}

func TestCLIBudgetList(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Monthly Food", "2000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("budget", "list")
	if err != nil {
		t.Fatalf("budget list: %v", err)
	}
	if !strings.Contains(stdout, "Monthly Food") {
		t.Errorf("expected 'Monthly Food' in list: %s", stdout)
	}
}

func TestCLIBudgetListJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Monthly Food", "2000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("--json", "budget", "list")
	if err != nil {
		t.Fatalf("budget list --json: %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 budget, got %d", len(results))
	}
}

func TestCLIBudgetListEmpty(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("budget", "list")
	if err != nil {
		t.Fatalf("budget list: %v", err)
	}
	if !strings.Contains(stdout, "No budgets") {
		t.Errorf("expected 'No budgets' in output: %s", stdout)
	}
}

func TestCLIBudgetCheck(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("budget", "check", "--all")
	if err != nil {
		t.Fatalf("budget check --all: %v", err)
	}
	if !strings.Contains(stdout, "Food") {
		t.Errorf("expected 'Food' in check output: %s", stdout)
	}
	if !strings.Contains(stdout, "ok") {
		t.Errorf("expected 'ok' in check output: %s", stdout)
	}
}

func TestCLIBudgetCheckJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("--json", "budget", "check", "--all")
	if err != nil {
		t.Fatalf("budget check --json: %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0]["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", results[0]["status"])
	}
}

func TestCLIBudgetCheckNoFlags(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "check")
	if !strings.Contains(stderr, "specify") {
		t.Errorf("expected 'specify' error, got: %s", stderr)
	}
}

func TestCLIBudgetEdit(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("budget", "edit", "1", "--amount", "2500000", "--notify", "75")
	if err != nil {
		t.Fatalf("budget edit: %v", err)
	}
	if !strings.Contains(stdout, "updated") {
		t.Errorf("expected 'updated' in output: %s", stdout)
	}
}

func TestCLIBudgetEditJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("--json", "budget", "edit", "1", "--amount", "2500000")
	if err != nil {
		t.Fatalf("budget edit --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIBudgetRm(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("budget", "rm", "1")
	if err != nil {
		t.Fatalf("budget rm: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output: %s", stdout)
	}
}

func TestCLIBudgetRmJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	stdout, _, err := cli.run("--json", "budget", "rm", "1")
	if err != nil {
		t.Fatalf("budget rm --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLIBudgetRmInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "rm", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIBudgetEditInvalidID(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "edit", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' error, got: %s", stderr)
	}
}

func TestCLIBudgetRmNotFound(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "rm", "99")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr)
	}
}

func TestCLIBudgetListAll(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Active", "1000000", "-c", "Restaurant", "--period", "monthly")
	_ = cli.svc.RemoveBudget(1)

	stdout, _, err := cli.run("budget", "list", "--all")
	if err != nil {
		t.Fatalf("budget list --all: %v", err)
	}
	if !strings.Contains(stdout, "Active") {
		t.Errorf("expected 'Active' in --all list: %s", stdout)
	}
	if !strings.Contains(stdout, "inactive") {
		t.Errorf("expected 'inactive' in --all list: %s", stdout)
	}
}

func TestCLIBudgetCheckNoResults(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "one_time", "--from", "2025-01-01", "--to", "2025-06-30")
	_ = cli.svc.RemoveBudget(1)

	stdout, _, err := cli.run("budget", "check", "--all")
	if err != nil {
		t.Fatalf("budget check --all: %v", err)
	}
	if !strings.Contains(stdout, "No budgets to check") {
		t.Errorf("expected 'No budgets to check': %s", stdout)
	}
}

func TestCLIBudgetEditMissing(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "edit", "99", "--amount", "2500000")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found', got: %s", stderr)
	}
}

func TestBudgetDisplayNameNull(t *testing.T) {
	nullBudget := &gen.Budget{Name: sql.NullString{Valid: false}}
	name := budgetDisplayName(nullBudget)
	if !strings.Contains(name, "id:") {
		t.Errorf("expected id-based name for null, got: %s", name)
	}

	validBudget := &gen.Budget{Name: sql.NullString{String: "Valid", Valid: true}}
	name = budgetDisplayName(validBudget)
	if name != "Valid" {
		t.Errorf("expected 'Valid', got: %s", name)
	}
}

func TestBudgetNotifyPctNull(t *testing.T) {
	nullBudget := &gen.Budget{NotifyAtPct: sql.NullInt64{Valid: false}}
	pct := budgetNotifyPct(nullBudget)
	if pct != 80 {
		t.Errorf("expected 80 for null notify, got %d", pct)
	}

	validBudget := &gen.Budget{NotifyAtPct: sql.NullInt64{Int64: 90, Valid: true}}
	pct = budgetNotifyPct(validBudget)
	if pct != 90 {
		t.Errorf("expected 90, got %d", pct)
	}
}

func TestCLIBudgetEditInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("budget", "set", "Food", "1000000", "-c", "Restaurant", "--period", "monthly")

	_, stderr, _ := cli.run("budget", "edit", "1", "--amount", "not-a-number")
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' for bad amount, got: %s", stderr)
	}
}

type budgetListErrorQuerier struct {
	gen.Querier
}

func (q *budgetListErrorQuerier) ListActiveBudgets(ctx context.Context) ([]*gen.Budget, error) {
	return nil, fmt.Errorf("mock budget list error")
}

type budgetCheckErrorQuerier struct {
	gen.Querier
}

func (q *budgetCheckErrorQuerier) ListActiveBudgets(ctx context.Context) ([]*gen.Budget, error) {
	return nil, fmt.Errorf("mock budget check error")
}

type budgetEditErrorQuerier struct {
	gen.Querier
}

func (q *budgetEditErrorQuerier) GetBudgetByID(ctx context.Context, id int64) (*gen.Budget, error) {
	return nil, fmt.Errorf("mock budget edit error")
}

func TestCLIBudgetListError(t *testing.T) {
	dbase, _ := db.Open(":memory:")
	_ = db.Migrate(dbase)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetListErrorQuerier{Querier: gen.New(dbase)})
	origOverride := getServiceOverride
	getServiceOverride = func(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
		return svc, dbase, nil
	}
	t.Cleanup(func() { getServiceOverride = origOverride })

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"budget", "list"})
	_ = cmd.Execute()

	if !strings.Contains(stderr.String(), "mock budget list error") || !strings.Contains(stderr.String(), "Error") {
		t.Errorf("expected error output, got stderr: %s", stderr.String())
	}
}

func TestCLIBudgetCheckError(t *testing.T) {
	dbase, _ := db.Open(":memory:")
	_ = db.Migrate(dbase)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetCheckErrorQuerier{Querier: gen.New(dbase)})
	origOverride := getServiceOverride
	getServiceOverride = func(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
		return svc, dbase, nil
	}
	t.Cleanup(func() { getServiceOverride = origOverride })

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"budget", "check", "--all"})
	_ = cmd.Execute()

	if !strings.Contains(stderr.String(), "mock budget check error") {
		t.Errorf("expected error output, got stderr: %s", stderr.String())
	}
}

func TestCLIBudgetEditError(t *testing.T) {
	dbase, _ := db.Open(":memory:")
	_ = db.Migrate(dbase)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetEditErrorQuerier{Querier: gen.New(dbase)})
	origOverride := getServiceOverride
	getServiceOverride = func(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
		return svc, dbase, nil
	}
	t.Cleanup(func() { getServiceOverride = origOverride })

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"budget", "edit", "1", "--name", "New"})
	_ = cmd.Execute()

	if !strings.Contains(stderr.String(), "mock budget edit error") {
		t.Errorf("expected error output, got stderr: %s", stderr.String())
	}
}

func TestCLIAddExpenseForeignCurrency(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	stdout, _, err := cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}
	if !strings.Contains(stdout, "10 USD") {
		t.Errorf("expected '10 USD' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "158.000") {
		t.Errorf("expected base amount in output, got: %s", stdout)
	}
}

func TestCLIAddExpenseForeignCurrencyJSON(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	stdout, _, err := cli.run("--json", "add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	if err != nil {
		t.Fatalf("add expense --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	currency, ok := result["currency"].(string)
	if !ok || currency != "USD" {
		t.Errorf("expected currency USD, got %v", result["currency"])
	}

	baseAmount, ok := result["base_amount"].(float64)
	if !ok {
		t.Fatalf("expected base_amount in JSON, got %T", result["base_amount"])
	}
	expectedBase := float64(10 * 15800)
	if baseAmount != expectedBase {
		t.Errorf("expected base_amount %f, got %f", expectedBase, baseAmount)
	}

	baseCurrency, ok := result["base_currency"].(string)
	if !ok || baseCurrency != "IDR" {
		t.Errorf("expected base_currency IDR, got %v", result["base_currency"])
	}
}

func TestCLIAddExpenseForeignCurrencyMissingRate(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, stderr, _ := cli.run("add", "expense", "50000", "Food", "-c", "Restaurant", "-a", "KRWAccount")
	if !strings.Contains(stderr, "wallet rate add KRW") {
		t.Errorf("expected actionable error, got: %s", stderr)
	}
}

func TestCLIListMultiCurrency(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	_, _, _ = cli.run("add", "expense", "50000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("list", "-n", "10")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "10 USD") {
		t.Errorf("expected '10 USD' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "158.000") {
		t.Errorf("expected base amount in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "50.000") {
		t.Errorf("expected IDR amount in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Base:") {
		t.Errorf("expected base total in output, got: %s", stdout)
	}
}

func TestCLIListMultiCurrencyJSON(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	_, _, _ = cli.run("add", "expense", "50000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("--json", "list", "-n", "10")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	transactions, ok := result["transactions"].([]interface{})
	if !ok {
		t.Fatalf("expected transactions array, got %T", result["transactions"])
	}

	baseTotal, ok := result["base_total"].(string)
	if !ok {
		t.Errorf("expected base_total in JSON, got %v", result["base_total"])
	} else if baseTotal == "" {
		t.Error("expected non-empty base_total")
	}

	baseCurrency, ok := result["base_currency"].(string)
	if !ok || baseCurrency != "IDR" {
		t.Errorf("expected base_currency IDR, got %v", result["base_currency"])
	}

	hasBaseAmount := false
	for _, tx := range transactions {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			continue
		}
		if ba, ok := txMap["base_amount"]; ok && ba != nil {
			hasBaseAmount = true
		}
	}

	if !hasBaseAmount {
		t.Error("expected at least one transaction with base_amount in JSON")
	}
	_ = len(transactions)
}

func TestCLIAddIncomeForeignCurrency(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	stdout, _, err := cli.run("add", "income", "100", "Freelance", "-c", "Freelance", "-a", "PaypalUSD")
	if err != nil {
		t.Fatalf("add income: %v", err)
	}
	if !strings.Contains(stdout, "100 USD") {
		t.Errorf("expected '100 USD' in output, got: %s", stdout)
	}
}

func TestCLIAddIncomeForeignCurrencyMissingRate(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, stderr, _ := cli.run("add", "income", "5000", "Income", "-c", "Salary", "-a", "JPYAccount")
	if !strings.Contains(stderr, "wallet rate add JPY") {
		t.Errorf("expected actionable error, got: %s", stderr)
	}
}

func TestCLIAddExpenseBaseCurrency(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}
	if !strings.Contains(stdout, "50000") {
		t.Errorf("expected 50000 in output, got: %s", stdout)
	}
}

func TestCLIListBaseCurrencyTransactionsNoBaseTotal(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("list", "-n", "10")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if strings.Contains(stdout, "Base:") {
		t.Errorf("expected no Base: total for base-only transactions, got: %s", stdout)
	}
}

func TestCLIListForeignCurrencyTransactionsOriginalOnly(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	_, _, _ = cli.run("add", "expense", "20", "Hotel", "-c", "Travel", "-a", "WiseEUR")

	stdout, _, err := cli.run("list", "-n", "10")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "10 USD") {
		t.Errorf("expected USD transaction, got: %s", stdout)
	}
	if !strings.Contains(stdout, "20 EUR") {
		t.Errorf("expected EUR transaction, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Base:") {
		t.Errorf("expected base total for mixed currencies, got: %s", stdout)
	}
}

func newMultiCurrencyTestCLI(t *testing.T) *testCLI {
	t.Helper()
	cleanup := setupTestService()
	t.Cleanup(cleanup)

	svc, _, _ := getServiceOverride(nil)
	_, _ = svc.CreateAccount("BCA", "checking", "IDR")
	_, _ = svc.CreateAccount("GoPay", "ewallet", "IDR")
	_, _ = svc.CreateAccount("WiseUSD", "checking", "USD")
	_, _ = svc.CreateAccount("WiseEUR", "checking", "EUR")
	_, _ = svc.CreateAccount("PaypalUSD", "checking", "USD")
	_, _ = svc.CreateAccount("KRWAccount", "checking", "KRW")
	_, _ = svc.CreateAccount("JPYAccount", "checking", "JPY")

	return &testCLI{t: t, svc: svc, cleanup: cleanup}
}
