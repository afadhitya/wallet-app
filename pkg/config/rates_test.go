package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRateConfig(t *testing.T) {
	cfg := DefaultRateConfig()

	if cfg.BaseCurrency != "IDR" {
		t.Errorf("expected base currency 'IDR', got '%s'", cfg.BaseCurrency)
	}
	if len(cfg.Rates) < 3 {
		t.Errorf("expected at least 3 default rates, got %d", len(cfg.Rates))
	}
	if cfg.Rates["USD"] <= 0 {
		t.Errorf("expected positive USD rate, got %d", cfg.Rates["USD"])
	}
}

func TestEnsureRatesFileCreatesDefault(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	if err := EnsureRatesFile(); err != nil {
		t.Fatalf("EnsureRatesFile: %v", err)
	}

	cfg, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}
	if cfg.BaseCurrency != "IDR" {
		t.Errorf("expected base currency 'IDR', got '%s'", cfg.BaseCurrency)
	}
}

func TestEnsureRatesFilePreservesExisting(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	if err := os.MkdirAll(filepath.Join(dir, ".config", "wallet"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := `base_currency = "USD"

[rates]
EUR = 11000
`
	path := filepath.Join(dir, ".config", "wallet", "rates.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := EnsureRatesFile(); err != nil {
		t.Fatalf("EnsureRatesFile: %v", err)
	}

	cfg, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}
	if cfg.BaseCurrency != "USD" {
		t.Errorf("expected base currency 'USD', got '%s'", cfg.BaseCurrency)
	}
	if cfg.Rates["EUR"] != 11000 {
		t.Errorf("expected EUR rate 11000, got %d", cfg.Rates["EUR"])
	}
}

func TestLoadRatesNonExistentFile(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	_, err := LoadRates()
	if err == nil {
		t.Fatal("expected error for missing rates file")
	}
	if !strings.Contains(err.Error(), "rate configuration not found") {
		t.Errorf("expected 'rate configuration not found' error, got: %v", err)
	}
}

func TestLoadRatesParseValid(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	if err := os.MkdirAll(filepath.Join(dir, ".config", "wallet"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := `base_currency = "USD"

[rates]
EUR = 11000
JPY = 105
`
	path := filepath.Join(dir, ".config", "wallet", "rates.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}
	if cfg.BaseCurrency != "USD" {
		t.Errorf("expected base currency 'USD', got '%s'", cfg.BaseCurrency)
	}
	if cfg.Rates["EUR"] != 11000 {
		t.Errorf("expected EUR rate 11000, got %d", cfg.Rates["EUR"])
	}
	if cfg.Rates["JPY"] != 105 {
		t.Errorf("expected JPY rate 105, got %d", cfg.Rates["JPY"])
	}
}

func TestSaveAndLoadRates(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	if err := os.MkdirAll(filepath.Join(dir, ".config", "wallet"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := `base_currency = "IDR"

[rates]
USD = 16000
`
	path := filepath.Join(dir, ".config", "wallet", "rates.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}

	cfg.Rates["SGD"] = 12000
	if err := SaveRates(cfg); err != nil {
		t.Fatalf("SaveRates: %v", err)
	}

	reloaded, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates after save: %v", err)
	}
	if reloaded.Rates["USD"] != 16000 {
		t.Errorf("expected USD 16000, got %d", reloaded.Rates["USD"])
	}
	if reloaded.Rates["SGD"] != 12000 {
		t.Errorf("expected SGD 12000, got %d", reloaded.Rates["SGD"])
	}
}

func TestSaveRatesCreatesDir(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	cfg := RateConfig{
		BaseCurrency: "EUR",
		Rates: map[string]int64{
			"USD": 10800,
		},
	}

	if err := SaveRates(cfg); err != nil {
		t.Fatalf("SaveRates: %v", err)
	}

	reloaded, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}
	if reloaded.BaseCurrency != "EUR" {
		t.Errorf("expected EUR, got '%s'", reloaded.BaseCurrency)
	}
}

func TestLoadRatesWithEmptyRatesMap(t *testing.T) {
	orig := userHomeDir
	dir := t.TempDir()
	userHomeDir = func() (string, error) { return dir, nil }
	defer func() { userHomeDir = orig }()

	if err := os.MkdirAll(filepath.Join(dir, ".config", "wallet"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := `base_currency = "GBP"
`
	path := filepath.Join(dir, ".config", "wallet", "rates.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadRates()
	if err != nil {
		t.Fatalf("LoadRates: %v", err)
	}
	if cfg.BaseCurrency != "GBP" {
		t.Errorf("expected GBP, got '%s'", cfg.BaseCurrency)
	}
	if cfg.Rates == nil {
		t.Fatal("expected non-nil rates map")
	}
}
