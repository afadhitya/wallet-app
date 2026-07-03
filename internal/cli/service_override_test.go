package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetService_ConfigIsDirectory(t *testing.T) {
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

	configPath := filepath.Join(tmpHome, ".config", "wallet", "config.toml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config path as directory: %v", err)
	}

	cmd := &cobra.Command{}
	_, _, err := getService(cmd)
	if err == nil {
		t.Fatal("expected error when config path is a directory")
	}
}

func TestWithService_ErrorPath(t *testing.T) {
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

	configPath := filepath.Join(tmpHome, ".config", "wallet", "config.toml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config path as directory: %v", err)
	}

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when getService fails")
	}
}

func TestInit_WithoutOverride(t *testing.T) {
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
	cmd.SetArgs([]string{"init"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init without override: %v", err)
	}
	if !strings.Contains(buf.String(), "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", buf.String())
	}
}
