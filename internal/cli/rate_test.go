package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/service"
)

func runTestCmd(args ...string) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func setupRateTest(t *testing.T) {
	t.Helper()

	service.SetTestRateConfig(service.TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})

	cleanup := setupTestService()
	t.Cleanup(func() {
		cleanup()
		service.ResetTestRateConfig()
	})
}

func TestCLIRateList(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("rate", "list")
	if err != nil {
		t.Fatalf("rate list: %v", err)
	}
	if !strings.Contains(stdout, "IDR") {
		t.Errorf("expected base currency in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "USD") {
		t.Errorf("expected USD in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "EUR") {
		t.Errorf("expected EUR in output, got: %s", stdout)
	}
}

func TestCLIRateListJSON(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("--json", "rate", "list")
	if err != nil {
		t.Fatalf("rate list --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["base_currency"] != "IDR" {
		t.Errorf("expected base_currency IDR, got %v", result["base_currency"])
	}
	rates, ok := result["rates"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected rates map in JSON, got %T", result["rates"])
	}
	if rates["USD"] == nil {
		t.Errorf("expected USD rate in JSON")
	}
	if rates["EUR"] == nil {
		t.Errorf("expected EUR rate in JSON")
	}
}

func TestCLIRateAdd(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("rate", "add", "KRW", "12")
	if err != nil {
		t.Fatalf("rate add: %v", err)
	}
	if !strings.Contains(stdout, "added") {
		t.Errorf("expected 'added' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "KRW") {
		t.Errorf("expected KRW in output, got: %s", stdout)
	}
}

func TestCLIRateAddJSON(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("--json", "rate", "add", "KRW", "12")
	if err != nil {
		t.Fatalf("rate add --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "added" {
		t.Errorf("expected status 'added', got %v", result["status"])
	}
}

func TestCLIRateSet(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("rate", "set", "USD", "16000")
	if err != nil {
		t.Fatalf("rate set: %v", err)
	}
	if !strings.Contains(stdout, "updated") {
		t.Errorf("expected 'updated' in output, got: %s", stdout)
	}
}

func TestCLIRateSetJSON(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("--json", "rate", "set", "USD", "16000")
	if err != nil {
		t.Fatalf("rate set --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "updated" {
		t.Errorf("expected status 'updated', got %v", result["status"])
	}
}

func TestCLIRateRm(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("rate", "rm", "EUR")
	if err != nil {
		t.Fatalf("rate rm: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", stdout)
	}
}

func TestCLIRateRmJSON(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("--json", "rate", "rm", "EUR")
	if err != nil {
		t.Fatalf("rate rm --json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

func TestCLIRateSetNonExistent(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "set", "KRW", "12")
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
	if !strings.Contains(err.Error(), "no existing rate") {
		t.Errorf("expected 'no existing rate' error, got: %v", err)
	}
}

func TestCLIRateAddInvalidRate(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "add", "KRW", "-5")
	if err == nil {
		t.Fatal("expected error for invalid rate")
	}
}

func TestCLIRateAddExisting(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "add", "USD", "16000")
	if err == nil {
		t.Fatal("expected error for existing rate")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCLIRateRmNonExistent(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "rm", "KRW")
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
	if !strings.Contains(err.Error(), "no configured rate") {
		t.Errorf("expected 'no configured rate' error, got: %v", err)
	}
}

func TestCLIRateListContainsInverse(t *testing.T) {
	setupRateTest(t)
	stdout, _, err := runTestCmd("rate", "list")
	if err != nil {
		t.Fatalf("rate list: %v", err)
	}
	if !strings.Contains(stdout, "Inverse") {
		t.Errorf("expected inverse in output, got: %s", stdout)
	}
}

func TestCLIRateListEmpty(t *testing.T) {
	cleanup := setupTestService()
	t.Cleanup(func() {
		cleanup()
		service.ResetTestRateConfig()
	})

	service.SetTestRateConfig(service.TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{},
	})

	stdout, _, err := runTestCmd("rate", "list")
	if err != nil {
		t.Fatalf("rate list: %v", err)
	}
	if !strings.Contains(stdout, "No exchange rates configured") {
		t.Errorf("expected empty message, got: %s", stdout)
	}
}

func TestCLIRateSetInvalidRate(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "set", "USD", "not-a-number")
	if err == nil {
		t.Fatal("expected error for non-numeric rate")
	}
}

func TestCLIRateAddInvalidRateNonNumeric(t *testing.T) {
	setupRateTest(t)
	_, _, err := runTestCmd("rate", "add", "KRW", "abc")
	if err == nil {
		t.Fatal("expected error for non-numeric rate")
	}
}
