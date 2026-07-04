package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCLIForecastDefaultText(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("forecast")
	if err != nil {
		t.Fatalf("forecast: %v", err)
	}
	if !strings.Contains(stdout, "Forecast: All Accounts") {
		t.Errorf("expected 'Forecast: All Accounts' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Start Balance") {
		t.Errorf("expected 'Start Balance' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Ending Balance") {
		t.Errorf("expected 'Ending Balance' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Netflix") {
		t.Errorf("expected 'Netflix' in output, got: %s", stdout)
	}
}

func TestCLIForecastMultiMonth(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("forecast", "--months", "3")
	if err != nil {
		t.Fatalf("forecast --months 3: %v", err)
	}
	if !strings.Contains(stdout, "3 months") {
		t.Errorf("expected '3 months' in output, got: %s", stdout)
	}
}

func TestCLIForecastAccountScoped(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("forecast", "--account", "BCA")
	if err != nil {
		t.Fatalf("forecast --account BCA: %v", err)
	}
	if !strings.Contains(stdout, "BCA") {
		t.Errorf("expected 'BCA' in output, got: %s", stdout)
	}
}

func TestCLIForecastUnknownAccount(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("forecast", "--account", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown account")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIForecastInvalidMonths(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("forecast", "--months", "0")
	if err == nil {
		t.Fatal("expected error for zero months")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIForecastEmptyState(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("forecast")
	if err != nil {
		t.Fatalf("forecast: %v", err)
	}
	if !strings.Contains(stdout, "No planned payments found") {
		t.Errorf("expected 'No planned payments found' in output, got: %s", stdout)
	}
}

func TestCLIForecastAccountJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("--json", "forecast", "--account", "BCA")
	if err != nil {
		t.Fatalf("forecast --json --account BCA: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	account, ok := result["account"].(map[string]interface{})
	if !ok {
		t.Error("expected 'account' key in JSON")
	}
	if account["name"] != "BCA" {
		t.Errorf("expected account name 'BCA', got %v", account["name"])
	}
}

func TestCLIForecastJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("--json", "forecast", "--months", "3")
	if err != nil {
		t.Fatalf("forecast --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if horizon, ok := result["horizon"].(float64); !ok || int(horizon) != 3 {
		t.Errorf("expected horizon 3, got %v", result["horizon"])
	}
	if _, ok := result["forecast"]; !ok {
		t.Error("expected 'forecast' key in JSON")
	}
	if _, ok := result["planned_payments"]; !ok {
		t.Error("expected 'planned_payments' key in JSON")
	}
	if _, ok := result["category_breakdown"]; !ok {
		t.Error("expected 'category_breakdown' key in JSON")
	}
}

func TestCLIForecastBillsDefaultText(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("forecast", "bills")
	if err != nil {
		t.Fatalf("forecast bills: %v", err)
	}
	if !strings.Contains(stdout, "Upcoming Bills") {
		t.Errorf("expected 'Upcoming Bills' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Running Total") {
		t.Errorf("expected 'Running Total' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Netflix") {
		t.Errorf("expected 'Netflix' in output, got: %s", stdout)
	}
}

func TestCLIForecastBillsCustomMonths(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("forecast", "bills", "--months", "2")
	if err != nil {
		t.Fatalf("forecast bills --months 2: %v", err)
	}
	if !strings.Contains(stdout, "2 months") {
		t.Errorf("expected '2 months' in output, got: %s", stdout)
	}
}

func TestCLIForecastBillsInvalidMonths(t *testing.T) {
	cli := newTestCLI(t)

	_, stderr, err := cli.run("forecast", "bills", "--months", "0")
	if err == nil {
		t.Fatal("expected error for zero months")
	}
	if !strings.Contains(stderr, "Error") {
		t.Errorf("expected error in stderr, got: %s", stderr)
	}
}

func TestCLIForecastBillsEmptyState(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("forecast", "bills")
	if err != nil {
		t.Fatalf("forecast bills: %v", err)
	}
	if !strings.Contains(stdout, "No planned payments found") {
		t.Errorf("expected 'No planned payments found' in output, got: %s", stdout)
	}
}

func TestCLIForecastBillsJSON(t *testing.T) {
	cli := newTestCLI(t)

	_, _, err := cli.run("bill", "add", "Netflix", "149000", "--monthly", "--day", "15", "-c", "Subscriptions", "-a", "BCA")
	if err != nil {
		t.Fatalf("bill add: %v", err)
	}

	stdout, _, err := cli.run("--json", "forecast", "bills")
	if err != nil {
		t.Fatalf("forecast bills --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if horizon, ok := result["horizon"].(float64); !ok || int(horizon) != 2 {
		t.Errorf("expected horizon 2, got %v", result["horizon"])
	}
	if _, ok := result["bills"]; !ok {
		t.Error("expected 'bills' key in JSON")
	}
	if _, ok := result["total_amount"]; !ok {
		t.Error("expected 'total_amount' key in JSON")
	}
	if _, ok := result["count"]; !ok {
		t.Error("expected 'count' key in JSON")
	}
}
