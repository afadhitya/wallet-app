package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
		"init", "add", "list", "category", "tag",
		"budget", "bill", "report", "forecast",
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
	testCases := []struct {
		name string
		args []string
	}{
		{"init", []string{"init"}},
		{"add", []string{"add"}},
		{"list", []string{"list"}},
		{"category", []string{"category"}},
		{"tag", []string{"tag"}},
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
	for _, sub := range []string{"init", "add", "list", "category", "tag", "budget", "bill", "report", "forecast"} {
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
	for _, name := range []string{"init", "add", "list"} {
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
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--json", "init"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected command to execute: %v", err)
	}

	got, err := cmd.Flags().GetBool("json")
	if err == nil && got {
		return
	}

	subCmd, _, err := cmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("expected to find init subcommand: %v", err)
	}

	got, err = subCmd.Flags().GetBool("json")
	if err != nil {
		t.Fatalf("expected --json flag on subcommand: %v", err)
	}
	if !got {
		t.Error("expected --json flag to be true on subcommand")
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
