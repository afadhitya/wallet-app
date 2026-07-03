package cli

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/pkg/config"
	"github.com/spf13/cobra"
)

func TestTagNames(t *testing.T) {
	tags := []*gen.Tag{
		{Name: "food"},
		{Name: "japan-2026"},
		{Name: "trip"},
	}
	names := tagNames(tags)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "food" {
		t.Errorf("expected 'food', got %q", names[0])
	}
	if names[1] != "japan-2026" {
		t.Errorf("expected 'japan-2026', got %q", names[1])
	}
	if names[2] != "trip" {
		t.Errorf("expected 'trip', got %q", names[2])
	}
}

func TestTagNames_Empty(t *testing.T) {
	names := tagNames(nil)
	if names == nil || len(names) != 0 {
		t.Errorf("expected nil/empty slice for nil input, got %v", names)
	}

	names = tagNames([]*gen.Tag{})
	if len(names) != 0 {
		t.Errorf("expected empty slice for empty input, got %d", len(names))
	}
}

func TestPrintErrJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	printErrJSON(buf, "something went wrong")

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if result["error"] != "something went wrong" {
		t.Errorf("expected error message 'something went wrong', got %q", result["error"])
	}
}

func TestFormatError_Nil(t *testing.T) {
	cmd := &cobra.Command{}
	err := formatError(cmd, nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestFormatError_NonJSON(t *testing.T) {
	cmd := &cobra.Command{}
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	err := formatError(cmd, fmt.Errorf("test failure"))
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(stderr.String(), "Error: test failure") {
		t.Errorf("expected stderr to contain error message, got %q", stderr.String())
	}
}

func TestFormatError_JSON(t *testing.T) {
	parent := &cobra.Command{}
	parent.PersistentFlags().Bool("json", false, "")
	cmd := &cobra.Command{}
	parent.AddCommand(cmd)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	_ = parent.ParseFlags([]string{"--json"})

	err := formatError(cmd, fmt.Errorf("json error test"))
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var result map[string]string
	if jsonErr := json.Unmarshal(stderr.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected JSON output in stderr, got: %s", stderr.String())
	}
	if result["error"] != "json error test" {
		t.Errorf("expected error 'json error test', got %q", result["error"])
	}
}

func TestFormatError_JSON_DirectFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	_ = cmd.Flags().Set("json", "true")
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	err := formatError(cmd, fmt.Errorf("direct flag test"))
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var result map[string]string
	if jsonErr := json.Unmarshal(stderr.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected JSON output in stderr, got: %s", stderr.String())
	}
	if result["error"] != "direct flag test" {
		t.Errorf("expected error 'direct flag test', got %q", result["error"])
	}
}

func TestExpandHomePath_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home dir: %v", err)
	}

	result := expandHomePath("~/test/path")
	expected := filepath.Join(home, "test", "path")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandHomePath_NoTilde(t *testing.T) {
	result := expandHomePath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expected '/absolute/path', got %q", result)
	}
}

func TestExpandHomePath_NoTildeNoSlash(t *testing.T) {
	result := expandHomePath("./relative/path")
	if result != "./relative/path" {
		t.Errorf("expected './relative/path', got %q", result)
	}
}

func TestIsJSON_DirectFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "json flag")
	_ = cmd.Flags().Set("json", "true")

	if !isJSON(cmd) {
		t.Error("expected isJSON to return true for direct flag")
	}
}

func TestIsJSON_DirectFlagFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "json flag")

	if isJSON(cmd) {
		t.Error("expected isJSON to return false when flag is false")
	}
}

func TestIsJSON_ParentPersistentFlag(t *testing.T) {
	parent := &cobra.Command{}
	parent.PersistentFlags().Bool("json", false, "")
	cmd := &cobra.Command{}
	parent.AddCommand(cmd)
	_ = parent.ParseFlags([]string{"--json"})

	if !isJSON(cmd) {
		t.Error("expected isJSON to return true from parent persistent flag")
	}
}

func TestIsJSON_NoFlagAnywhere(t *testing.T) {
	cmd := &cobra.Command{}

	if isJSON(cmd) {
		t.Error("expected isJSON to return false with no flag defined")
	}
}

func TestIsJSON_ParentPersistentFlagFalse(t *testing.T) {
	parent := &cobra.Command{}
	parent.PersistentFlags().Bool("json", false, "")
	cmd := &cobra.Command{}
	parent.AddCommand(cmd)

	if isJSON(cmd) {
		t.Error("expected isJSON to return false when parent flag is false")
	}
}

func TestPrintJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	data := map[string]interface{}{"status": "ok", "id": 42}
	if err := printJSON(buf, data); err != nil {
		t.Fatalf("printJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", result["status"])
	}
	if result["id"].(float64) != 42 {
		t.Errorf("expected id 42, got %v", result["id"])
	}
}

func TestGetService_WithTempHome(t *testing.T) {
	oldOverride := getServiceOverride
	getServiceOverride = nil
	defer func() { getServiceOverride = oldOverride }()

	tmpHome := t.TempDir()

	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer func() {
		if origHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", origHome)
		}
	}()

	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"list"})

	svc, db, err := getService(cmd)
	if err != nil {
		t.Fatalf("getService failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	accounts, err := svc.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts failed: %v", err)
	}
	if accounts == nil {
		t.Error("expected non-nil accounts list")
	}

	dbPath := filepath.Join(tmpHome, ".local", "share", "wallet", "wallet.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected database file to exist at %s", dbPath)
	}
}

func TestWithService_NoOverride_Success(t *testing.T) {
	oldOverride := getServiceOverride
	getServiceOverride = nil
	defer func() { getServiceOverride = oldOverride }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer func() {
		if origHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", origHome)
		}
	}()

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if !strings.Contains(buf.String(), "No transactions") {
		t.Errorf("expected 'No transactions', got %q", buf.String())
	}
}

func TestWithService_NoOverride_JSON(t *testing.T) {
	oldOverride := getServiceOverride
	getServiceOverride = nil
	defer func() { getServiceOverride = oldOverride }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer func() {
		if origHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", origHome)
		}
	}()

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--json", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list --json command failed: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
}

func TestResolveOut(t *testing.T) {
	cmd := &cobra.Command{}
	stdout, stderr := resolveOut(cmd)
	if stdout == nil {
		t.Error("expected non-nil stdout")
	}
	if stderr == nil {
		t.Error("expected non-nil stderr")
	}
}

func TestFormatError_PropagatesError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetErr(new(bytes.Buffer))

	original := errors.New("original error")
	result := formatError(cmd, original)
	if result != original {
		t.Errorf("expected same error to be returned, got %v", result)
	}
}

func TestGetService_ConfigLoadError(t *testing.T) {
	oldLoad := svcConfigLoad
	defer func() { svcConfigLoad = oldLoad }()
	svcConfigLoad = func(path string) (config.Config, error) {
		return config.Config{}, fmt.Errorf("mock load error")
	}

	cmd := &cobra.Command{}
	_, _, err := getService(cmd)
	if err == nil {
		t.Fatal("expected config load error")
	}
}

func TestGetService_MkdirAllError(t *testing.T) {
	oldMkdir := svcMkdirAll
	defer func() { svcMkdirAll = oldMkdir }()
	svcMkdirAll = func(path string, perm os.FileMode) error {
		return fmt.Errorf("mock mkdir error")
	}

	cmd := &cobra.Command{}
	_, _, err := getService(cmd)
	if err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestGetService_DBOpenError(t *testing.T) {
	oldOpen := svcDBOpen
	defer func() { svcDBOpen = oldOpen }()
	svcDBOpen = func(path string) (*sql.DB, error) {
		return nil, fmt.Errorf("mock open error")
	}

	cmd := &cobra.Command{}
	_, _, err := getService(cmd)
	if err == nil {
		t.Fatal("expected db open error")
	}
}

func TestGetService_MigrateError(t *testing.T) {
	oldMigrate := svcDBMigrate
	defer func() { svcDBMigrate = oldMigrate }()
	svcDBMigrate = func(db *sql.DB) error {
		return fmt.Errorf("mock migrate error")
	}

	cmd := &cobra.Command{}
	_, _, err := getService(cmd)
	if err == nil {
		t.Fatal("expected migrate error")
	}
}

func TestExpandHomePath_UserHomeDirError(t *testing.T) {
	origHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", origHome) }()

	result := expandHomePath("~/test/path")
	if result != "~/test/path" {
		t.Errorf("expected '~/test/path' when home dir fails, got %q", result)
	}
}
