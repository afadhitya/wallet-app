package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Database.Path != "~/.local/share/wallet/wallet.db" {
		t.Errorf("expected database path '~/.local/share/wallet/wallet.db', got '%s'", cfg.Database.Path)
	}
	if cfg.Display.Currency != "IDR" {
		t.Errorf("expected currency 'IDR', got '%s'", cfg.Display.Currency)
	}
	if cfg.Display.DateFormat != "2006-01-02" {
		t.Errorf("expected date format '2006-01-02', got '%s'", cfg.Display.DateFormat)
	}
	if cfg.Display.FirstDayOfWeek != "monday" {
		t.Errorf("expected first day of week 'monday', got '%s'", cfg.Display.FirstDayOfWeek)
	}
	if !cfg.AI.JSON {
		t.Errorf("expected AI JSON output enabled by default")
	}
	if cfg.Defaults.Account != "" {
		t.Errorf("expected default account to be empty, got '%s'", cfg.Defaults.Account)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/to/config.toml")
	if err != nil {
		t.Fatalf("Load() should not error on non-existent file: %v", err)
	}
	expected := DefaultConfig()
	if cfg.Database.Path != expected.Database.Path {
		t.Errorf("expected database path '%s', got '%s'", expected.Database.Path, cfg.Database.Path)
	}
}

func TestLoadTOMLOverride(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	content := `
[database]
path = "/custom/path/wallet.db"

[display]
currency = "USD"
date_format = "02-Jan-2006"
first_day_of_week = "sunday"

[ai]
json = false

[defaults]
account = "main"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Path != "/custom/path/wallet.db" {
		t.Errorf("expected database path '/custom/path/wallet.db', got '%s'", cfg.Database.Path)
	}
	if cfg.Display.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", cfg.Display.Currency)
	}
	if cfg.Display.DateFormat != "02-Jan-2006" {
		t.Errorf("expected date format '02-Jan-2006', got '%s'", cfg.Display.DateFormat)
	}
	if cfg.Display.FirstDayOfWeek != "sunday" {
		t.Errorf("expected first day of week 'sunday', got '%s'", cfg.Display.FirstDayOfWeek)
	}
	if cfg.AI.JSON {
		t.Errorf("expected AI JSON output disabled")
	}
	if cfg.Defaults.Account != "main" {
		t.Errorf("expected default account 'main', got '%s'", cfg.Defaults.Account)
	}
}

func TestLoadPartialOverride(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	content := `
[display]
currency = "JPY"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Display.Currency != "JPY" {
		t.Errorf("expected currency 'JPY', got '%s'", cfg.Display.Currency)
	}
	if cfg.Display.DateFormat != "2006-01-02" {
		t.Errorf("expected default date format '2006-01-02', got '%s'", cfg.Display.DateFormat)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home dir: %v", err)
	}
	expectedDBPath := filepath.Join(home, ".local", "share", "wallet", "wallet.db")
	if cfg.Database.Path != expectedDBPath {
		t.Errorf("expected database path '%s', got '%s'", expectedDBPath, cfg.Database.Path)
	}
}

func TestLoadEmptyPathDefaultsToXDG(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error = %v", err)
	}
	if cfg.Database.Path != "~/.local/share/wallet/wallet.db" {
		t.Errorf("expected default database path, got '%s'", cfg.Database.Path)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home dir: %v", err)
	}

	expanded := expandPath("~/test/path")
	expected := filepath.Join(home, "test", "path")
	if expanded != expected {
		t.Errorf("expected expanded path '%s', got '%s'", expected, expanded)
	}
}

func TestExpandPathNoPrefix(t *testing.T) {
	result := expandPath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expected unchanged '/absolute/path', got '%s'", result)
	}
}

func TestExpandPathInConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	content := `
[database]
path = "~/wallet/data/test.db"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home dir: %v", err)
	}

	expected := filepath.Join(home, "wallet", "data", "test.db")
	if cfg.Database.Path != expected {
		t.Errorf("expected expanded path '%s', got '%s'", expected, cfg.Database.Path)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	if err := os.WriteFile(configPath, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
	if !strings.Contains(err.Error(), "parse config file") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestLoadEmptyPathHomeDirError(t *testing.T) {
	orig := userHomeDir
	userHomeDir = func() (string, error) {
		return "", errors.New("no home directory")
	}
	defer func() { userHomeDir = orig }()

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when home directory lookup fails")
	}
	if !strings.Contains(err.Error(), "home directory") {
		t.Errorf("expected 'home directory' error, got: %v", err)
	}
}

func TestExpandPathHomeDirError(t *testing.T) {
	orig := userHomeDir
	userHomeDir = func() (string, error) {
		return "", errors.New("no home directory")
	}
	defer func() { userHomeDir = orig }()

	result := expandPath("~/test/path")
	if result != "~/test/path" {
		t.Errorf("expected unchanged path '~/test/path', got '%s'", result)
	}
}

func TestLoadReadFileError(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error when loading a directory as config")
	}
	if !strings.Contains(err.Error(), "read config file") {
		t.Errorf("expected 'read config file' error, got: %v", err)
	}
}

func TestLoadHomeDirInExpandPathError(t *testing.T) {
	orig := userHomeDir
	userHomeDir = func() (string, error) {
		return "", errors.New("no home directory")
	}
	defer func() { userHomeDir = orig }()

	cfg, _ := Load("~/test/config.toml")
	if cfg.Database.Path != "~/.local/share/wallet/wallet.db" {
		t.Errorf("expected default database path, got '%s'", cfg.Database.Path)
	}
}
