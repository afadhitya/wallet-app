package cli

import (
	"strings"
	"testing"
)

func TestCLIAccountIntegration_AddAndList(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("account", "add", "Mandiri", "--type", "savings")
	if err != nil {
		t.Fatalf("account add: %v", err)
	}
	if !strings.Contains(stdout, "Mandiri") {
		t.Errorf("expected 'Mandiri' in output, got: %s", stdout)
	}

	stdout, _, err = cli.run("account", "list")
	if err != nil {
		t.Fatalf("account list: %v", err)
	}
	if !strings.Contains(stdout, "Mandiri") {
		t.Errorf("expected 'Mandiri' in list, got: %s", stdout)
	}
	if !strings.Contains(stdout, "BCA") {
		t.Errorf("expected 'BCA' in list, got: %s", stdout)
	}
	if !strings.Contains(stdout, "GoPay") {
		t.Errorf("expected 'GoPay' in list, got: %s", stdout)
	}
}

func TestCLIAccountIntegration_AddEditArchive(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("account", "add", "Mandiri", "--type", "savings")
	if err != nil {
		t.Fatalf("account add: %v", err)
	}
	if !strings.Contains(stdout, "Mandiri") {
		t.Errorf("expected 'Mandiri' in output, got: %s", stdout)
	}

	stdout, _, err = cli.run("account", "edit", "3", "--name", "Mandiri Baru")
	if err != nil {
		t.Fatalf("account edit: %v", err)
	}
	if !strings.Contains(stdout, "Account 3 updated") {
		t.Errorf("expected edit success, got: %s", stdout)
	}

	stdout, _, err = cli.run("account", "archive", "3", "--force")
	if err != nil {
		t.Fatalf("account archive: %v", err)
	}
	if !strings.Contains(stdout, "Account 3 archived") {
		t.Errorf("expected archive success, got: %s", stdout)
	}

	stdout, _, err = cli.run("account", "list")
	if err != nil {
		t.Fatalf("account list: %v", err)
	}
	if strings.Contains(stdout, "Mandiri") {
		t.Errorf("archived account should not appear in default list")
	}

	stdout, _, err = cli.run("account", "list", "--all")
	if err != nil {
		t.Fatalf("account list --all: %v", err)
	}
	if !strings.Contains(stdout, "Mandiri Baru") {
		t.Errorf("archived account should appear with --all, got: %s", stdout)
	}
}

func TestCLIAccountIntegration_JSONWorkflow(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("--json", "account", "add", "ShopeePay", "--type", "ewallet")
	if err != nil {
		t.Fatalf("account add --json: %v", err)
	}
	result := extractJSONData(t, stdout)
	if result["name"] != "ShopeePay" {
		t.Errorf("expected name 'ShopeePay', got %v", result["name"])
	}

	stdout, _, err = cli.run("--json", "account", "list")
	if err != nil {
		t.Fatalf("account list --json: %v", err)
	}
	arr := extractJSONArray(t, stdout)
	if len(arr) != 3 {
		t.Errorf("expected 3 accounts, got %d", len(arr))
	}

	stdout, _, err = cli.run("--json", "account", "edit", "3", "--name", "ShopeePay Updated")
	if err != nil {
		t.Fatalf("account edit --json: %v", err)
	}
	result = extractJSONData(t, stdout)
	if result["name"] != "ShopeePay Updated" {
		t.Errorf("expected name 'ShopeePay Updated', got %v", result["name"])
	}

	stdout, _, err = cli.run("--json", "account", "archive", "3", "--force")
	if err != nil {
		t.Fatalf("account archive --json: %v", err)
	}
	result = extractJSONData(t, stdout)
	if result["status"] != "archived" {
		t.Errorf("expected status 'archived', got %v", result["status"])
	}
}

func TestCLIAccountIntegration_EditType(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("account", "edit", "1", "--type", "savings")
	if err != nil {
		t.Fatalf("account edit --type: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 updated") {
		t.Errorf("expected edit success, got: %s", stdout)
	}

	stdout, _, err = cli.run("account", "list")
	if err != nil {
		t.Fatalf("account list: %v", err)
	}
	if !strings.Contains(stdout, "savings") {
		t.Errorf("expected 'savings' type in list, got: %s", stdout)
	}
}

func TestCLIAccountIntegration_EditSortOrder(t *testing.T) {
	cli := newTestCLI(t)

	stdout, _, err := cli.run("account", "edit", "1", "--sort-order", "10")
	if err != nil {
		t.Fatalf("account edit --sort-order: %v", err)
	}
	if !strings.Contains(stdout, "Account 1 updated") {
		t.Errorf("expected edit success, got: %s", stdout)
	}
}
