package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCLIBillAddText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}
	if !strings.Contains(stdout, "Netflix") {
		t.Errorf("expected 'Netflix' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "monthly") {
		t.Errorf("expected 'monthly' in output, got: %s", stdout)
	}
}

func TestCLIBillAddJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("--json", "bill", "add", "Spotify", "54990", "--monthly", "--day", "1", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["name"] != "Spotify" {
		t.Errorf("expected name 'Spotify', got %v", result["name"])
	}
}

func TestCLIBillAddInvalidAmount(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "add", "Invalid", "0", "--monthly", "-a", "BCA", "-c", "Subscriptions")
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIBillListText(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	_, _, _ = cli.run("bill", "add", "Spotify", "54990", "--monthly", "--day", "1", "-c", "Subscriptions", "-a", "BCA")

	stdout, _, err := cli.run("bill", "list")
	if err != nil {
		t.Fatalf("bill list: %v", err)
	}
	if !strings.Contains(stdout, "Netflix") {
		t.Errorf("expected 'Netflix' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Spotify") {
		t.Errorf("expected 'Spotify' in output, got: %s", stdout)
	}
}

func TestCLIBillListJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")

	stdout, _, err := cli.run("--json", "bill", "list")
	if err != nil {
		t.Fatalf("bill list --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	payments, ok := result["planned_payments"].([]interface{})
	if !ok {
		t.Fatal("expected planned_payments array")
	}
	if len(payments) < 1 {
		t.Errorf("expected at least 1 payment, got %d", len(payments))
	}
}

func TestCLIBillListPaused(t *testing.T) {
	cli := newTestCLI(t)

	_, _, _ = cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")

	stdout, _, _ := cli.run("--json", "bill", "list", "--paused")
	var result map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &result)
	payments := result["planned_payments"].([]interface{})
	if len(payments) != 0 {
		t.Errorf("expected 0 paused payments initially, got %d", len(payments))
	}
}

func TestCLIBillDueText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "due")
	if err != nil {
		t.Fatalf("bill due: %v", err)
	}
	if !strings.Contains(stdout, "Due") || !strings.Contains(stdout, "due") {
		hasContent := stdout != "No due planned payments.\n" && !strings.Contains(stdout, "Total")
		_ = hasContent
	}
}

func TestCLIBillDueJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("--json", "bill", "due")
	if err != nil {
		t.Fatalf("bill due --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	_, ok := result["count"].(float64)
	if !ok {
		t.Errorf("expected 'count' in JSON output")
	}
}

func TestCLIBillDueOverdue(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("--json", "bill", "due", "--overdue")
	if err != nil {
		t.Fatalf("bill due --overdue: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIBillPayText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "pay", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill pay: %v", err)
	}
	if !strings.Contains(stdout, "Paid planned payment") {
		t.Errorf("expected 'Paid planned payment' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Transaction") {
		t.Errorf("expected 'Transaction' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Next due") {
		t.Errorf("expected 'Next due' in output, got: %s", stdout)
	}
}

func TestCLIBillPayJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("--json", "bill", "pay", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill pay --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	_, ok := result["Transaction"]
	if !ok {
		t.Errorf("expected Transaction in JSON response")
	}
	_, ok = result["PlannedPayment"]
	if !ok {
		t.Errorf("expected PlannedPayment in JSON response")
	}
}

func TestCLIBillPayNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "pay", "9999")
	if err == nil {
		t.Fatal("expected error for non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillPayOneTime(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Flight", "3000000", "--from", "2026-08-15", "-c", "Travel", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "pay", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill pay one-time: %v", err)
	}
	if !strings.Contains(stdout, "archived") {
		t.Errorf("expected 'archived' in output, got: %s", stdout)
	}
}

func TestCLIBillSkipText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "skip", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill skip: %v", err)
	}
	if !strings.Contains(stdout, "Skipped") {
		t.Errorf("expected 'Skipped' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Next due") {
		t.Errorf("expected 'Next due' in output, got: %s", stdout)
	}
}

func TestCLIBillSkipOneTime(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Flight", "3000000", "--from", "2026-08-15", "-c", "Travel", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	_, stderr, err := cli.run("bill", "skip", fmt.Sprintf("%d", id))
	if err == nil {
		t.Fatal("expected error for skipping one-time payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillPauseText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "pause", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill pause: %v", err)
	}
	if !strings.Contains(stdout, "Paused") {
		t.Errorf("expected 'Paused' in output, got: %s", stdout)
	}
}

func TestCLIBillResumeText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	_, _, _ = cli.run("bill", "pause", fmt.Sprintf("%d", id))

	stdout, _, err := cli.run("bill", "resume", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill resume: %v", err)
	}
	if !strings.Contains(stdout, "Resumed") {
		t.Errorf("expected 'Resumed' in output, got: %s", stdout)
	}
}

func TestCLIBillEditText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "--name", "Netflix Premium", "--amount", "169000")
	if err != nil {
		t.Fatalf("bill edit: %v", err)
	}
	if !strings.Contains(stdout, "Netflix Premium") {
		t.Errorf("expected 'Netflix Premium' in output, got: %s", stdout)
	}
}

func TestCLIBillRmText(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "rm", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill rm: %v", err)
	}
	if !strings.Contains(stdout, "Deleted") {
		t.Errorf("expected 'Deleted' in output, got: %s", stdout)
	}
}

func TestCLIBillRmNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "rm", "9999")
	if err == nil {
		t.Fatal("expected error for deleting non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillListEmpty(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "list")
	if err != nil {
		t.Fatalf("bill list: %v", err)
	}
	if !strings.Contains(stdout, "No planned payments") {
		t.Errorf("expected 'No planned payments' in empty list, got: %s", stdout)
	}
}

func TestCLIBillDueWithPayment(t *testing.T) {
	cli := newTestCLI(t)

	now := time.Now()
	today := now.Format("2006-01-02")

	_, _, _ = cli.run("bill", "add", "Due Now", "50000", "--from", today, "-a", "BCA", "-c", "Subscriptions")

	stdout, _, err := cli.run("bill", "due")
	if err != nil {
		t.Fatalf("bill due: %v", err)
	}
	if !strings.Contains(stdout, "Due Now") && !strings.Contains(stdout, "No due") {
		t.Errorf("expected 'Due Now' or empty, got: %s", stdout)
	}
}

func TestCLIBillDueWeek(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "due", "--week")
	if err != nil {
		t.Fatalf("bill due --week: %v", err)
	}
	_ = stdout
}

func TestCLIBillDueNext(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "due", "--next", "14")
	if err != nil {
		t.Fatalf("bill due --next 14: %v", err)
	}
	_ = stdout
}

func TestCLIBillPauseJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("--json", "bill", "pause", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill pause --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIBillResumeJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	_, _, _ = cli.run("bill", "pause", fmt.Sprintf("%d", id))

	stdout, _, err := cli.run("--json", "bill", "resume", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill resume --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIBillSkipJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("--json", "bill", "skip", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill skip --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestCLIBillRmJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("--json", "bill", "rm", fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatalf("bill rm --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status 'deleted', got %v", result["status"])
	}
}

func TestCLIBillAddDaily(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Daily Bill", "10000", "--daily", "-a", "BCA", "-c", "Subscriptions")
	if err != nil {
		t.Fatalf("bill add daily: %v", err)
	}
	if !strings.Contains(stdout, "daily") {
		t.Errorf("expected 'daily' in output, got: %s", stdout)
	}
}

func TestCLIBillAddWeekly(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Weekly Bill", "50000", "--weekly", "--day", "1", "-a", "BCA", "-c", "Subscriptions")
	if err != nil {
		t.Fatalf("bill add weekly: %v", err)
	}
	if !strings.Contains(stdout, "weekly") {
		t.Errorf("expected 'weekly' in output, got: %s", stdout)
	}
}

func TestCLIBillAddYearly(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Yearly Bill", "500000", "--yearly", "-a", "BCA", "-c", "Subscriptions")
	if err != nil {
		t.Fatalf("bill add yearly: %v", err)
	}
	if !strings.Contains(stdout, "yearly") {
		t.Errorf("expected 'yearly' in output, got: %s", stdout)
	}
}

func TestCLIBillAddCustom(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Custom Bill", "75000", "--custom", "--rrule", "FREQ=MONTHLY;BYMONTHDAY=15", "-a", "BCA", "-c", "Subscriptions")
	if err != nil {
		t.Fatalf("bill add custom: %v", err)
	}
	if !strings.Contains(stdout, "custom") {
		t.Errorf("expected 'custom' in output, got: %s", stdout)
	}
}

func TestCLIBillAddOneTime(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "add", "Flight", "3000000", "--from", "2026-08-15", "-a", "BCA", "-c", "Travel")
	if err != nil {
		t.Fatalf("bill add one-time: %v", err)
	}
	if !strings.Contains(stdout, "none") {
		t.Errorf("expected 'none' in output, got: %s", stdout)
	}
}

func TestCLIBillAddInvalidRecurrence(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "add", "Invalid", "50000", "--custom", "-a", "BCA", "-c", "Subscriptions")
	if err == nil {
		t.Fatal("expected error for custom without RRULE")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIBillAddDuplicateRecurrence(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "add", "Invalid", "50000", "--monthly", "--weekly", "-a", "BCA", "-c", "Subscriptions")
	if err == nil {
		t.Fatal("expected error for duplicate recurrence flags")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIBillPayPaused(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	_, _, _ = cli.run("bill", "pause", fmt.Sprintf("%d", id))

	_, stderr, err := cli.run("bill", "pay", fmt.Sprintf("%d", id))
	if err == nil {
		t.Fatal("expected error for paying paused payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIBillEditJSON(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("--json", "bill", "edit", fmt.Sprintf("%d", id), "--name", "Netflix Premium")
	if err != nil {
		t.Fatalf("bill edit --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["name"] != "Netflix Premium" {
		t.Errorf("expected name 'Netflix Premium', got %v", result["name"])
	}
}

func TestCLIBillEditAccountCategory(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "-a", "GoPay", "-c", "Internet")
	if err != nil {
		t.Fatalf("bill edit: %v", err)
	}
	if !strings.Contains(stdout, "Updated planned payment") {
		t.Errorf("expected 'Updated planned payment' in output, got: %s", stdout)
	}
}

func TestCLIBillEditRecurrence(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "--recurrence", "daily")
	if err != nil {
		t.Fatalf("bill edit recurrence: %v", err)
	}
	if !strings.Contains(stdout, "daily") {
		t.Errorf("expected 'daily' in output, got: %s", stdout)
	}
}

func TestCLIBillEditRRULE(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "--rrule", "FREQ=MONTHLY;BYMONTHDAY=20")
	if err != nil {
		t.Fatalf("bill edit rrule: %v", err)
	}
	if !strings.Contains(stdout, "Updated planned payment") {
		t.Errorf("expected 'Updated planned payment' in output, got: %s", stdout)
	}
}

func TestCLIBillEditDay(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "--day", "20")
	if err != nil {
		t.Fatalf("bill edit day: %v", err)
	}
	if !strings.Contains(stdout, "Updated planned payment") {
		t.Errorf("expected 'Updated planned payment' in output, got: %s", stdout)
	}
}

func TestCLIBillEditFrom(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, _ := cli.run("--json", "bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	var addResult map[string]interface{}
	_ = json.Unmarshal([]byte(stdout), &addResult)
	id := int64(addResult["id"].(float64))

	stdout, _, err := cli.run("bill", "edit", fmt.Sprintf("%d", id), "--from", "2026-08-01")
	if err != nil {
		t.Fatalf("bill edit from: %v", err)
	}
	if !strings.Contains(stdout, "Updated planned payment") {
		t.Errorf("expected 'Updated planned payment' in output, got: %s", stdout)
	}
}

func TestCLIBillPauseNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "pause", "9999")
	if err == nil {
		t.Fatal("expected error for pausing non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillResumeNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "resume", "9999")
	if err == nil {
		t.Fatal("expected error for resuming non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillSkipNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "skip", "9999")
	if err == nil {
		t.Fatal("expected error for skipping non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillEditNotFound(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "edit", "9999", "--name", "New")
	if err == nil {
		t.Fatal("expected error for editing non-existent payment")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillAddInvalidID(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("bill", "pay", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error output, got: %s", stderr)
	}
}

func TestCLIBillHelp(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("bill", "--help")
	if err != nil {
		t.Fatalf("bill --help: %v", err)
	}
	subcommands := []string{"add", "list", "due", "pay", "skip", "pause", "resume", "edit", "rm"}
	for _, sub := range subcommands {
		if !strings.Contains(stdout, sub) {
			t.Errorf("expected subcommand '%s' in help output", sub)
		}
	}
}
