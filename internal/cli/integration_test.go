package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

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

func extractJSONData(t *testing.T, stdout string) map[string]interface{} {
	t.Helper()
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("unmarshal JSON envelope: %v, output: %s", err, stdout)
	}
	data, ok := envelope["data"]
	if !ok {
		t.Fatalf("envelope missing 'data' field: %s", stdout)
	}
	m, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be an object, got type %T: %s", data, stdout)
	}
	return m
}

func extractJSONArray(t *testing.T, stdout string) []interface{} {
	t.Helper()
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("unmarshal JSON envelope: %v, output: %s", err, stdout)
	}
	data, ok := envelope["data"]
	if !ok {
		t.Fatalf("envelope missing 'data' field: %s", stdout)
	}
	arr, ok := data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be an array, got type %T: %s", data, stdout)
	}
	return arr
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

	result := extractJSONData(t, stdout)
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

	arr := extractJSONArray(t, stdout)
	if len(arr) < 10 {
		t.Errorf("expected at least 10 categories, got %d", len(arr))
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

	result := extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
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

	arr := extractJSONArray(t, stdout)
	if len(arr) != 1 {
		t.Errorf("expected 1 tag, got %d", len(arr))
	}
}

func TestCLITagRmJSON(t *testing.T) {
	cli := newTestCLI(t)
	_, _, _ = cli.run("tag", "add", "temp")

	stdout, _, err := cli.run("--json", "tag", "rm", "temp")
	if err != nil {
		t.Fatalf("tag rm --json: %v", err)
	}

	_ = extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
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

	_ = extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
	if result["name"] != "Monthly Food" {
		t.Errorf("expected name 'Monthly Food', got %v", result["name"])
	}
	if result["period"] != "monthly" {
		t.Errorf("expected period 'monthly', got %v", result["period"])
	}
}

func TestCLIBudgetSetAllCategories(t *testing.T) {
	cli := newTestCLI(t)
	stdout, _, err := cli.run("--json", "budget", "set", "All my money", "2000000", "-A", "--period", "monthly")
	if err != nil {
		t.Fatalf("budget set --json: %v", err)
	}

	result := extractJSONData(t, stdout)
	if result["name"] != "All my money" {
		t.Errorf("expected name 'All my money', got %v", result["name"])
	}
	if result["period"] != "monthly" {
		t.Errorf("expected period 'monthly', got %v", result["period"])
	}
	if result["all_categories"] != true {
		t.Errorf("expected all_categories 'true', got %v", result["all_categories"])
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

func TestCLIBudgetSetAllCategoriesConflict(t *testing.T) {
	cli := newTestCLI(t)
	_, stderr, _ := cli.run("budget", "set", "All my money", "2000000", "-c", "Restaurant", "--period", "monthly", "-A")
	if !strings.Contains(stderr, "none of the others can be") {
		t.Errorf("expected 'none of the others can be' error, got: %s", stderr)
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

	result := extractJSONData(t, stdout)
	budgets, ok := result["budgets"].([]interface{})
	if !ok {
		t.Fatalf("expected budgets array, got %T", result["budgets"])
	}
	if len(budgets) != 1 {
		t.Errorf("expected 1 budget, got %d", len(budgets))
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

	result := extractJSONData(t, stdout)
	budgets, ok := result["budgets"].([]interface{})
	if !ok {
		t.Fatalf("expected budgets array, got %T", result["budgets"])
	}
	if len(budgets) != 1 {
		t.Errorf("expected 1 result, got %d", len(budgets))
	}
	results0 := budgets[0].(map[string]interface{})
	if results0["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", results0["status"])
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

	_ = extractJSONData(t, stdout)
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

	result := extractJSONData(t, stdout)
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
	logger := testLogger()
	dbase, _ := db.Open(":memory:", logger)
	_ = db.Migrate(dbase, logger)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetListErrorQuerier{Querier: gen.New(dbase)}, logger)
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
	logger := testLogger()
	dbase, _ := db.Open(":memory:", logger)
	_ = db.Migrate(dbase, logger)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetCheckErrorQuerier{Querier: gen.New(dbase)}, logger)
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
	logger := testLogger()
	dbase, _ := db.Open(":memory:", logger)
	_ = db.Migrate(dbase, logger)
	t.Cleanup(func() { _ = dbase.Close() })
	svc := service.NewWithQuerier(dbase, &budgetEditErrorQuerier{Querier: gen.New(dbase)}, logger)
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

	result := extractJSONData(t, stdout)

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

	result := extractJSONData(t, stdout)

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

func TestCLIReportBaseCurrency(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")
	_, _, _ = cli.run("add", "income", "200000", "Freelance", "-c", "Freelance", "-a", "BCA")

	stdout, _, err := cli.run("report")
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	if !strings.Contains(stdout, "Report —") {
		t.Errorf("expected report header, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Income:") {
		t.Errorf("expected income total, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Expenses:") {
		t.Errorf("expected expense total, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Coffee & Snacks") {
		t.Errorf("expected category in breakdown, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Transactions:") {
		t.Errorf("expected transaction count, got: %s", stdout)
	}
}

func TestCLIReportJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "income", "200000", "Freelance", "-c", "Freelance", "-a", "BCA")

	stdout, _, err := cli.run("--json", "report")
	if err != nil {
		t.Fatalf("report --json: %v", err)
	}

	result := extractJSONData(t, stdout)

	if result["base_currency"] != "IDR" {
		t.Errorf("expected base_currency IDR, got %v", result["base_currency"])
	}

	income, ok := result["income_total"].(float64)
	if !ok || int64(income) != 200000 {
		t.Errorf("expected income_total 200000, got %v", result["income_total"])
	}

	expense, ok := result["expense_total"].(float64)
	if !ok || int64(expense) != 50000 {
		t.Errorf("expected expense_total 50000, got %v", result["expense_total"])
	}

	cats, ok := result["income_categories"].([]interface{})
	if !ok {
		cats, ok = result["expense_categories"].([]interface{})
	}
	if !ok || len(cats) == 0 {
		t.Errorf("expected categories in JSON")
	}
}

func TestCLIReportMixedCurrency(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	_, _, _ = cli.run("add", "expense", "20", "Hotel", "-c", "Travel", "-a", "WiseEUR")

	stdout, _, err := cli.run("report")
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	if !strings.Contains(stdout, "Base Currency: IDR") {
		t.Errorf("expected base currency IDR, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Coffee & Snacks") {
		t.Errorf("expected category in breakdown, got: %s", stdout)
	}
}

func TestCLIReportMixedCurrencyJSON(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")
	_, _, _ = cli.run("add", "income", "100", "Payment", "-c", "Freelance", "-a", "PaypalUSD")

	stdout, _, err := cli.run("--json", "report", "--by", "account")
	if err != nil {
		t.Fatalf("report --json --by account: %v", err)
	}

	result := extractJSONData(t, stdout)

	accts, ok := result["by_account"].([]interface{})
	if !ok || len(accts) == 0 {
		t.Errorf("expected by_account in JSON, got: %s", stdout)
	}
}

func TestCLIReportExcludesAdjustment(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("adjust", "BCA", "0", "Correction")
	_, _, _ = cli.run("add", "income", "200000", "Salary", "-c", "Salary", "-a", "BCA")

	stdout, _, err := cli.run("report")
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	if !strings.Contains(stdout, "Income:") {
		t.Errorf("expected income section, got: %s", stdout)
	}
}

func TestCLIReportNoTransactions(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("report")
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	if !strings.Contains(stdout, "No transactions found") {
		t.Errorf("expected empty message, got: %s", stdout)
	}
}

func TestCLIReportWithMonthFilter(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA", "-d", "2026-06-15")
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("report", "--month", "june")
	if err != nil {
		t.Fatalf("report --month: %v", err)
	}

	if !strings.Contains(stdout, "50.000") {
		t.Errorf("expected June transaction, got: %s", stdout)
	}
	if strings.Contains(stdout, "35.000") {
		t.Errorf("expected no July transaction, got: %s", stdout)
	}
}

func TestCLIReportWithAccountFilter(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "expense", "10", "AWS", "-c", "Subscriptions", "-a", "WiseUSD")

	stdout, _, err := cli.run("report", "--account", "BCA")
	if err != nil {
		t.Fatalf("report --account: %v", err)
	}

	if !strings.Contains(stdout, "Coffee & Snacks") {
		t.Errorf("expected BCA category, got: %s", stdout)
	}
	if strings.Contains(stdout, "Subscriptions") {
		t.Errorf("expected no WiseUSD category, got: %s", stdout)
	}
}

func TestCLIReportByCategory(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "expense", "35000", "Lunch", "-c", "Restaurant", "-a", "BCA")

	stdout, _, err := cli.run("report", "--by", "category")
	if err != nil {
		t.Fatalf("report --by category: %v", err)
	}

	if !strings.Contains(stdout, "Expenses by Category") {
		t.Errorf("expected 'Expenses by Category', got: %s", stdout)
	}
	if !strings.Contains(stdout, "TOTAL") {
		t.Errorf("expected TOTAL row, got: %s", stdout)
	}
}

func TestCLIReportByAccount(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")
	_, _, _ = cli.run("add", "income", "100000", "Salary", "-c", "Salary", "-a", "GoPay")

	stdout, _, err := cli.run("report", "--by", "account")
	if err != nil {
		t.Fatalf("report --by account: %v", err)
	}

	if !strings.Contains(stdout, "Transactions by Account") {
		t.Errorf("expected 'Transactions by Account', got: %s", stdout)
	}
	if !strings.Contains(stdout, "TOTAL") {
		t.Errorf("expected TOTAL row, got: %s", stdout)
	}
}

func TestCLIReportByTag(t *testing.T) {
	cli := newTestCLI(t)

	workTag, _ := cli.svc.CreateTag("work")

	txn, _ := cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = cli.svc.AddTransactionTag(txn.Transaction.ID, workTag.ID)
	_, _ = cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 30000, Description: "Snack", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})

	stdout, _, err := cli.run("report", "--by", "tag", "--month", "july")
	if err != nil {
		t.Fatalf("report --by tag: %v", err)
	}

	if !strings.Contains(stdout, "Expenses by Tag") {
		t.Errorf("expected 'Expenses by Tag', got: %s", stdout)
	}
	if !strings.Contains(stdout, "work") {
		t.Errorf("expected work tag in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "(untagged)") {
		t.Errorf("expected untagged row, got: %s", stdout)
	}
}

func TestCLIReportByCategoryJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")

	stdout, _, err := cli.run("--json", "report", "--by", "category")
	if err != nil {
		t.Fatalf("report --json --by category: %v", err)
	}

	result := extractJSONData(t, stdout)
	if _, ok := result["by_category"]; !ok {
		t.Errorf("expected by_category in JSON, got: %s", stdout)
	}
}

func TestCLIReportByAccountJSON(t *testing.T) {
	cli := newMultiCurrencyTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")

	stdout, _, err := cli.run("--json", "report", "--by", "account")
	if err != nil {
		t.Fatalf("report --json --by account: %v", err)
	}

	result := extractJSONData(t, stdout)
	if _, ok := result["by_account"]; !ok {
		t.Errorf("expected by_account in JSON, got: %s", stdout)
	}
}

func TestCLIReportByTagJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("add", "expense", "50000", "Coffee", "-c", "Coffee & Snacks", "-a", "BCA")

	stdout, _, err := cli.run("--json", "report", "--by", "tag")
	if err != nil {
		t.Fatalf("report --json --by tag: %v", err)
	}

	result := extractJSONData(t, stdout)
	if _, ok := result["by_tag"]; !ok {
		t.Errorf("expected by_tag in JSON, got: %s", stdout)
	}
}

func TestCLIReportInvalidBy(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("report", "--by", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid --by")
	}
}

func TestCLIReportInvalidExport(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("report", "--export", "pdf")
	if err == nil {
		t.Fatal("expected error for invalid --export")
	}
}

func TestCLIReportExportCSV(t *testing.T) {
	cli := newTestCLI(t)

	workTag, _ := cli.svc.CreateTag("work")

	txn, _ := cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})
	_ = cli.svc.AddTransactionTag(txn.Transaction.ID, workTag.ID)

	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test-export.csv"

	stdout, _, err := cli.run("report", "--export", "csv", "--output", outputPath, "--month", "july")
	if err != nil {
		t.Fatalf("report --export csv: %v", err)
	}

	if !strings.Contains(stdout, "Exported to:") {
		t.Errorf("expected 'Exported to:' in output, got: %s", stdout)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading exported CSV: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "date,type,amount,currency,base_amount,category,account,description,tags") {
		t.Errorf("expected CSV header, got: %s", content)
	}
	if !strings.Contains(content, "2026-07-01,expense,50000,IDR,,") {
		t.Errorf("expected transaction row, got: %s", content)
	}
	if !strings.Contains(content, "Coffee & Snacks,BCA,Coffee,work") {
		t.Errorf("expected full row with tags, got: %s", content)
	}
}

func TestCLIReportExportCSVDefaultFilename(t *testing.T) {
	cli := newTestCLI(t)

	_, _ = cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	stdout, _, err := cli.run("report", "--export", "csv", "--month", "2026-07")
	if err != nil {
		t.Fatalf("report --export csv: %v", err)
	}

	if !strings.Contains(stdout, "wallet-report-2026-07.csv") {
		t.Errorf("expected default filename, got: %s", stdout)
	}

	_ = os.Remove("wallet-report-2026-07.csv")
}

func TestCLIReportExportCSVJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _ = cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-01",
	})

	tmpDir := t.TempDir()
	outputPath := tmpDir + "/json-export.csv"

	stdout, _, err := cli.run("--json", "report", "--export", "csv", "--output", outputPath, "--month", "july")
	if err != nil {
		t.Fatalf("report --json --export csv: %v", err)
	}

	result := extractJSONData(t, stdout)
	if _, ok := result["file_path"]; !ok {
		t.Errorf("expected file_path in JSON, got: %s", stdout)
	}
	if _, ok := result["rows"]; !ok {
		t.Errorf("expected rows in JSON, got: %s", stdout)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected CSV file to exist at %s: %v", outputPath, err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSV file")
	}
}

func TestCLIReportInvalidMonth(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("report", "--month", "not-a-month")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestCLIReportInvalidMonthJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("--json", "report", "--month", "not-a-month")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}

	if !strings.Contains(stderr, `"success": false`) {
		t.Errorf("expected JSON error in stderr, got: %s", stderr)
	}
}

func TestCLIReportDateRangeOverride(t *testing.T) {
	cli := newTestCLI(t)

	_, _ = cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 50000, Description: "Coffee", Category: "Coffee & Snacks",
		Account: "BCA", Date: "2026-07-15",
	})
	_, _ = cli.svc.AddExpense(service.CreateExpenseParams{
		Amount: 35000, Description: "Lunch", Category: "Restaurant",
		Account: "BCA", Date: "2026-07-01",
	})

	stdout, _, err := cli.run("report", "--month", "2026-07", "--from", "2026-07-10", "--to", "2026-07-20")
	if err != nil {
		t.Fatalf("report with date range: %v", err)
	}

	if !strings.Contains(stdout, "50.000") {
		t.Errorf("expected July 15 transaction, got: %s", stdout)
	}
	if strings.Contains(stdout, "35.000") {
		t.Errorf("expected no July 1 transaction, got: %s", stdout)
	}
}
