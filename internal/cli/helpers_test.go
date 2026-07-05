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
	"github.com/afadhitya/wallet-app/internal/service"
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

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(stderr.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected JSON output in stderr, got: %s", stderr.String())
	}

	if success, ok := result["success"].(bool); !ok || success {
		t.Errorf("expected success false, got %v", result["success"])
	}

	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object, got %T", result["error"])
	}
	if errorObj["message"] != "json error test" {
		t.Errorf("expected error message 'json error test', got %q", errorObj["message"])
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

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(stderr.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected JSON output in stderr, got: %s", stderr.String())
	}

	if success, ok := result["success"].(bool); !ok || success {
		t.Errorf("expected success false, got %v", result["success"])
	}

	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object, got %T", result["error"])
	}
	if errorObj["message"] != "direct flag test" {
		t.Errorf("expected error message 'direct flag test', got %q", errorObj["message"])
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

func TestNewSuccessResponse(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	parent := &cobra.Command{Use: "wallet"}
	parent.AddCommand(cmd)

	data := map[string]interface{}{"key": "value"}
	resp := newSuccessResponse(data, cmd)

	if !resp.Success {
		t.Error("expected success true")
	}

	if resp.Data == nil {
		t.Error("expected non-nil data")
	}

	m, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", resp.Data)
	}
	if m["key"] != "value" {
		t.Errorf("expected data.key 'value', got %v", m["key"])
	}

	if resp.Meta.Command == "" {
		t.Error("expected non-empty meta.command")
	}

	if resp.Meta.Timestamp == "" {
		t.Error("expected non-empty meta.timestamp")
	}
}

func TestClassifyErrorNotFound(t *testing.T) {
	testCases := []struct {
		err         error
		expected    string
		description string
	}{
		{&service.NotFoundError{Entity: "category", Name: "food"}, ErrCodeCategoryNotFound, "category not found"},
		{&service.NotFoundError{Entity: "account", Name: "bca"}, ErrCodeAccountNotFound, "account not found"},
		{&service.NotFoundError{Entity: "tag", Name: "vip"}, ErrCodeTagNotFound, "tag not found"},
		{&service.NotFoundError{Entity: "transaction", Name: "1"}, ErrCodeTransactionNotFound, "transaction not found"},
		{&service.NotFoundError{Entity: "budget", Name: "1"}, ErrCodeBudgetNotFound, "budget not found"},
		{&service.NotFoundError{Entity: "planned payment", Name: "1"}, ErrCodePlannedPaymentNotFound, "bill not found"},
		{&service.NotFoundError{Entity: "unknown", Name: "x"}, ErrCodeNotFound, "unknown not found"},
	}

	for _, tc := range testCases {
		code, _ := classifyError(tc.err)
		if code != tc.expected {
			t.Errorf("%s: expected %s, got %s", tc.description, tc.expected, code)
		}
	}
}

func TestClassifyErrorValidation(t *testing.T) {
	tests := []struct {
		err     error
		code    string
		suggest string
	}{
		{&service.ValidationError{Field: "amount", Message: "invalid"}, ErrCodeInvalidAmount, "invalid"},
		{&service.ValidationError{Field: "date", Message: "bad date"}, ErrCodeInvalidDate, "bad date"},
		{&service.ValidationError{Field: "start_date", Message: "bad"}, ErrCodeInvalidDate, "bad"},
		{&service.ValidationError{Field: "state", Message: "paused"}, ErrCodeBillPaused, "unpause the planned payment first"},
		{&service.ValidationError{Field: "state", Message: "not paused"}, ErrCodeBillPaused, "planned payment is not paused"},
		{&service.ValidationError{Field: "state", Message: "already archived"}, ErrCodeValidation, "already archived"},
		{&service.ValidationError{Field: "unknown_field", Message: "fail"}, ErrCodeValidation, "fail"},
	}

	for _, tc := range tests {
		code, suggestion := classifyError(tc.err)
		if code != tc.code {
			t.Errorf("expected code %s, got %s", tc.code, code)
		}
		if suggestion != tc.suggest {
			t.Errorf("expected suggestion %q, got %q", tc.suggest, suggestion)
		}
	}
}

func TestClassifyErrorSentinel(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{service.ErrInvalidAmount, ErrCodeInvalidAmount},
		{service.ErrRateConfigMissing, ErrCodeExchangeRateConfig},
		{service.ErrRateMustBePositive, ErrCodeExchangeRateInvalid},
		{service.ErrDuplicateName, ErrCodeValidation},
		{service.ErrNotFound, ErrCodeNotFound},
		{service.ErrMissingField, ErrCodeValidation},
	}

	for _, tc := range tests {
		code, _ := classifyError(tc.err)
		if code != tc.code {
			t.Errorf("expected %s for %v, got %s", tc.code, tc.err, code)
		}
	}
}

func TestClassifyErrorRateNotFound(t *testing.T) {
	err := &service.RateNotFoundError{Currency: "KRW", Base: "IDR"}
	code, suggestion := classifyError(err)
	if code != ErrCodeExchangeRateNotFound {
		t.Errorf("expected %s, got %s", ErrCodeExchangeRateNotFound, code)
	}
	if !strings.Contains(suggestion, "wallet rate add KRW") {
		t.Errorf("expected actionable suggestion, got %s", suggestion)
	}
}

func TestClassifyErrorDBError(t *testing.T) {
	err := fmt.Errorf("database: sql error occurred")
	code, _ := classifyError(err)
	if code != ErrCodeDBError {
		t.Errorf("expected %s, got %s", ErrCodeDBError, code)
	}
}

func TestClassifyErrorInvalidInput(t *testing.T) {
	err := fmt.Errorf("invalid something required")
	code, _ := classifyError(err)
	if code != ErrCodeInvalidInput {
		t.Errorf("expected %s, got %s", ErrCodeInvalidInput, code)
	}
}

func TestClassifyErrorInternal(t *testing.T) {
	err := fmt.Errorf("something unexpected happened")
	code, _ := classifyError(err)
	if code != ErrCodeInternal {
		t.Errorf("expected %s, got %s", ErrCodeInternal, code)
	}
}

func TestPrintSuccessJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	parent := &cobra.Command{Use: "wallet"}
	parent.AddCommand(cmd)

	buf := new(bytes.Buffer)
	if err := printSuccessJSON(buf, map[string]int{"id": 42}, cmd); err != nil {
		t.Fatalf("printSuccessJSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	success, _ := result["success"].(bool)
	if !success {
		t.Error("expected success true")
	}

	data := result["data"].(map[string]interface{})
	if data["id"].(float64) != 42 {
		t.Errorf("expected data.id 42, got %v", data["id"])
	}

	meta := result["meta"].(map[string]interface{})
	if meta["command"] == nil {
		t.Error("expected meta.command")
	}
	if meta["timestamp"] == nil {
		t.Error("expected meta.timestamp")
	}
}

func TestPrintErrorJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := printErrorJSON(buf, "TEST_CODE", "test message", "try harder"); err != nil {
		t.Fatalf("printErrorJSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	success, _ := result["success"].(bool)
	if success {
		t.Error("expected success false")
	}

	errObj := result["error"].(map[string]interface{})
	if errObj["code"] != "TEST_CODE" {
		t.Errorf("expected code TEST_CODE, got %v", errObj["code"])
	}
	if errObj["message"] != "test message" {
		t.Errorf("expected message, got %v", errObj["message"])
	}
	if errObj["suggestion"] != "try harder" {
		t.Errorf("expected suggestion, got %v", errObj["suggestion"])
	}
}

func TestPrintErrorJSONNoSuggestion(t *testing.T) {
	buf := new(bytes.Buffer)
	_ = printErrorJSON(buf, "CODE", "msg", "")

	var result map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &result)

	errObj := result["error"].(map[string]interface{})
	if _, exists := errObj["suggestion"]; exists {
		t.Error("expected no suggestion field when empty")
	}
}

func TestSkillFileExists(t *testing.T) {
	skillPath := filepath.Join("..", "..", "skill", "SKILL.md")
	info, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("skill/SKILL.md not found at %s: %v", skillPath, err)
	}
	if info.IsDir() {
		t.Error("skill/SKILL.md should be a file, not a directory")
	}
}

func TestSkillFileContent(t *testing.T) {
	skillPath := filepath.Join("..", "..", "skill", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill/SKILL.md: %v", err)
	}

	text := string(content)
	casesSensitive := []string{
		"name: wallet",
		"wallet add expense",
		"wallet bill due",
		"wallet forecast",
		"--json",
		"success",
		"data",
		"CATEGORY_NOT_FOUND",
		"BILL_PAUSED",
	}

	for _, check := range casesSensitive {
		if !strings.Contains(text, check) {
			t.Errorf("skill/SKILL.md missing expected content: %q", check)
		}
	}

	if !strings.Contains(strings.ToLower(text), "auto-create") {
		t.Error("skill/SKILL.md missing guidance about auto-creating tags")
	}
}
