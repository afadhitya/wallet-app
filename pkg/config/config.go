package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database DatabaseConfig `toml:"database"`
	Display  DisplayConfig  `toml:"display"`
	AI       AIConfig       `toml:"ai"`
	Defaults DefaultsConfig `toml:"defaults"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type DisplayConfig struct {
	Currency       string `toml:"currency"`
	DateFormat     string `toml:"date_format"`
	FirstDayOfWeek string `toml:"first_day_of_week"`
}

type AIConfig struct {
	JSON bool `toml:"json"`
}

type DefaultsConfig struct {
	Account string `toml:"account"`
}

func DefaultConfig() Config {
	return Config{
		Database: DatabaseConfig{
			Path: "~/.local/share/wallet/wallet.db",
		},
		Display: DisplayConfig{
			Currency:       "IDR",
			DateFormat:     "2006-01-02",
			FirstDayOfWeek: "monday",
		},
		AI: AIConfig{
			JSON: true,
		},
		Defaults: DefaultsConfig{
			Account: "",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, fmt.Errorf("home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "wallet", "config.toml")
	} else {
		path = expandPath(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config file %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config file %s: %w", path, err)
	}

	cfg.Database.Path = expandPath(cfg.Database.Path)
	return cfg, nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
