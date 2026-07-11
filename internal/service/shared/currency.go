package shared

import (
	"fmt"
	"math"

	"github.com/afadhitya/wallet-app/pkg/config"
)

var (
	ErrRateConfigMissing  = fmt.Errorf("rate configuration not found (run 'wallet init')")
	ErrRateMustBePositive = fmt.Errorf("rate must be a positive integer")

	LoadRates = config.LoadRates
	SaveRates = config.SaveRates
)

type TestRateConfig struct {
	BaseCurrency string
	Rates        map[string]int64
}

func SetTestRateConfig(cfg TestRateConfig) {
	LoadRates = func() (config.RateConfig, error) {
		rates := make(map[string]int64)
		for k, v := range cfg.Rates {
			rates[k] = v
		}
		return config.RateConfig{BaseCurrency: cfg.BaseCurrency, Rates: rates}, nil
	}
	SaveRates = func(rc config.RateConfig) error {
		cfg.BaseCurrency = rc.BaseCurrency
		cfg.Rates = rc.Rates
		return nil
	}
}

func ResetTestRateConfig() {
	LoadRates = config.LoadRates
	SaveRates = config.SaveRates
}

type RateNotFoundError struct {
	Currency string
	Base     string
}

func (e *RateNotFoundError) Error() string {
	return fmt.Sprintf("no configured rate for %s → %s (use 'wallet rate add %s <rate>')", e.Currency, e.Base, e.Currency)
}

type RateInfo struct {
	Currency string
	Rate     int64
	Inverse  float64
}

func loadRateConfig() (config.RateConfig, error) {
	cfg, err := LoadRates()
	if err != nil {
		if err.Error() == "rate configuration not found" {
			return config.RateConfig{}, ErrRateConfigMissing
		}
		return config.RateConfig{}, fmt.Errorf("load rate config: %w", err)
	}
	return cfg, nil
}

func GetBaseCurrency() (string, error) {
	cfg, err := loadRateConfig()
	if err != nil {
		return "", err
	}
	return cfg.BaseCurrency, nil
}

func GetRate(currency string) (int64, error) {
	cfg, err := loadRateConfig()
	if err != nil {
		return 0, err
	}

	if currency == cfg.BaseCurrency {
		return 1, nil
	}

	rate, ok := cfg.Rates[currency]
	if !ok {
		return 0, &RateNotFoundError{Currency: currency, Base: cfg.BaseCurrency}
	}

	if rate <= 0 {
		return 0, fmt.Errorf("%w for %s", ErrRateMustBePositive, currency)
	}

	return rate, nil
}

func Convert(amount int64, fromCurrency string) (int64, error) {
	cfg, err := loadRateConfig()
	if err != nil {
		return 0, err
	}

	if fromCurrency == cfg.BaseCurrency {
		return amount, nil
	}

	rate, ok := cfg.Rates[fromCurrency]
	if !ok {
		return 0, &RateNotFoundError{Currency: fromCurrency, Base: cfg.BaseCurrency}
	}

	if rate <= 0 {
		return 0, fmt.Errorf("%w for %s", ErrRateMustBePositive, fromCurrency)
	}

	result := int64(math.Round(float64(amount) * float64(rate)))
	return result, nil
}

func ListRates() (string, map[string]int64, error) {
	cfg, err := loadRateConfig()
	if err != nil {
		return "", nil, err
	}
	return cfg.BaseCurrency, cfg.Rates, nil
}

func AddRate(currency string, rate int64) error {
	if rate <= 0 {
		return fmt.Errorf("%w: %d", ErrRateMustBePositive, rate)
	}

	cfg, err := loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		return fmt.Errorf("cannot add a rate for the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; exists {
		return fmt.Errorf("rate for %s already exists (use 'wallet rate set %s <rate>' to update)", currency, currency)
	}

	cfg.Rates[currency] = rate
	return SaveRates(cfg)
}

func SetRate(currency string, rate int64) error {
	if rate <= 0 {
		return fmt.Errorf("%w: %d", ErrRateMustBePositive, rate)
	}

	cfg, err := loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		return fmt.Errorf("cannot set a rate for the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; !exists {
		return fmt.Errorf("no existing rate for %s (use 'wallet rate add %s <rate>' to add)", currency, currency)
	}

	cfg.Rates[currency] = rate
	return SaveRates(cfg)
}

func RemoveRate(currency string) error {
	cfg, err := loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		return fmt.Errorf("cannot remove the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; !exists {
		return fmt.Errorf("no configured rate for %s", currency)
	}

	delete(cfg.Rates, currency)
	return SaveRates(cfg)
}
