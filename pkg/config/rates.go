package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type RateConfig struct {
	BaseCurrency string           `toml:"base_currency"`
	Rates        map[string]int64 `toml:"rates"`
}

func DefaultRateConfig() RateConfig {
	return RateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"SGD": 11800,
			"EUR": 17200,
			"JPY": 105,
			"MYR": 3400,
		},
	}
}

func ratesConfigPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".config", "wallet", "rates.toml"), nil
}

func LoadRates() (RateConfig, error) {
	cfg := DefaultRateConfig()

	path, err := ratesConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, fmt.Errorf("rate configuration not found at %s (run 'wallet init')", path)
		}
		return cfg, fmt.Errorf("read rates file %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse rates file %s: %w", path, err)
	}

	if cfg.Rates == nil {
		cfg.Rates = make(map[string]int64)
	}

	return cfg, nil
}

func EnsureRatesFile() error {
	path, err := ratesConfigPath()
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(path); statErr == nil {
		return nil
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("stat rates file %s: %w", path, statErr)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	cfg := DefaultRateConfig()
	if err := SaveRates(cfg); err != nil {
		return err
	}

	return nil
}

func SaveRates(cfg RateConfig) error {
	path, err := ratesConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create rates file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("write rates file %s: %w", path, err)
	}

	return nil
}
