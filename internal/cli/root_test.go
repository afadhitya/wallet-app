package cli

import (
	"bytes"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func setupTestService() func() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dbase, err := db.Open(":memory:", logger)
	if err != nil {
		panic(err)
	}
	if err := db.Migrate(dbase, logger); err != nil {
		_ = dbase.Close()
		panic(err)
	}
	svc := service.New(dbase, logger)
	getServiceOverride = func(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
		return svc, dbase, nil
	}

	service.SetTestRateConfig(service.TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})

	return func() {
		getServiceOverride = nil
		service.ResetTestRateConfig()
		_ = dbase.Close()
	}
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}
	if cmd.Use != "wallet" {
		t.Errorf("expected Use 'wallet', got '%s'", cmd.Use)
	}
}

func TestRootCmdExecutesWithoutError(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command should execute without error: %v", err)
	}
}

func TestSubcommandRegistration(t *testing.T) {
	cmd := NewRootCmd()
	expectedSubcommands := []string{
		"init", "add", "list", "edit", "rm",
		"category", "tag", "account", "adjust",
		"budget", "bill", "report", "forecast",
		"rate", "docs", "version", "update",
	}

	subcommands := cmd.Commands()
	subNames := make(map[string]bool)
	for _, sub := range subcommands {
		subNames[sub.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !subNames[name] {
			t.Errorf("expected subcommand '%s' to be registered", name)
		}
	}

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}
}

func TestJSONFlag(t *testing.T) {
	cmd := NewRootCmd()
	flag := cmd.PersistentFlags().Lookup("json")
	if flag == nil {
		t.Fatal("expected --json persistent flag to exist")
	}
	if flag.Name != "json" {
		t.Errorf("expected flag name 'json', got '%s'", flag.Name)
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got '%s'", flag.DefValue)
	}
}

func TestJSONFlagAvailable(t *testing.T) {
	cleanup := setupTestService()
	defer cleanup()

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"--json", "list"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected command with --json flag to execute: %v", err)
	}

	jsonVal, err := cmd.Flags().GetBool("json")
	if err != nil {
		t.Fatalf("expected to get --json flag: %v", err)
	}
	if !jsonVal {
		t.Error("expected --json flag to be true")
	}
}

func TestSubcommandExecution(t *testing.T) {
	cleanup := setupTestService()
	defer cleanup()

	testCases := []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"category", []string{"category", "list"}},
		{"tag", []string{"tag", "list"}},

		{"budget", []string{"budget"}},
		{"bill", []string{"bill"}},
		{"report", []string{"report"}},
		{"forecast", []string{"forecast"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("%s command failed: %v", tc.name, err)
			}
		})
	}
}

func TestRootCmdHelpOutput(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	for _, sub := range []string{"init", "add", "list", "category", "tag", "account", "edit", "rm", "adjust", "budget", "bill", "rate", "report", "forecast"} {
		if !strings.Contains(output, sub) {
			t.Errorf("expected help output to contain '%s'", sub)
		}
	}
}

func TestRootCmdHelpFlag(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"-h"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help flag failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("expected help output to contain 'Usage:'")
	}
}

func TestSubcommandHelp(t *testing.T) {
	cleanup := setupTestService()
	defer cleanup()

	for _, name := range []string{"init", "add", "list", "edit", "rm", "category", "tag", "adjust"} {
		t.Run(name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{name, "--help"})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("%s --help failed: %v", name, err)
			}
		})
	}
}

func TestInvalidSubcommand(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent subcommand")
	}
}

func TestJSONFlagPersistsToSubcommand(t *testing.T) {
	cleanup := setupTestService()
	defer cleanup()

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--json", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected command to execute: %v", err)
	}

	got, err := cmd.Flags().GetBool("json")
	if err == nil && got {
		return
	}

	subCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("expected to find list subcommand: %v", err)
	}

	got, err = subCmd.Flags().GetBool("json")
	if err != nil {
		t.Fatalf("expected --json flag on subcommand: %v", err)
	}
	if !got {
		t.Error("expected --json flag to be true on subcommand")
	}
}

func TestDocsMarkdownGeneratesFiles(t *testing.T) {
	root := NewRootCmd()

	dir, err := os.MkdirTemp("", "wallet-docs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"docs", "markdown", "--output", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("docs markdown failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected generated markdown files, got none")
	}

	found := map[string]bool{}
	for _, e := range entries {
		found[e.Name()] = true
	}

	for _, name := range []string{"wallet.md", "wallet_add.md", "wallet_list.md"} {
		if !found[name] {
			t.Errorf("expected generated file %s not found", name)
		}
	}

	hiddenFiles := []string{"wallet_docs.md", "wallet_docs_markdown.md"}
	for _, name := range hiddenFiles {
		if found[name] {
			t.Errorf("hidden command file %s should not have been generated", name)
		}
	}
}

func TestDocsCommandHidden(t *testing.T) {
	root := NewRootCmd()

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "docs") {
		t.Error("docs command should not appear in help output")
	}
}

func TestDocsMarkdownDefaultOutput(t *testing.T) {
	root := NewRootCmd()

	dir, err := os.MkdirTemp("", "wallet-docs-default-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	root.SetArgs([]string{"docs", "markdown", "--output", dir})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))

	if err := root.Execute(); err != nil {
		t.Fatalf("docs markdown failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected generated markdown files, got none")
	}
}

func init() {
	for _, cmd := range []*cobra.Command{
		NewRootCmd(),
	} {
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
	}
}
