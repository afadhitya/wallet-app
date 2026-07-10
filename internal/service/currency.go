package service

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/afadhitya/wallet-app/pkg/config"
)

var (
	ErrRateConfigMissing  = fmt.Errorf("rate configuration not found (run 'wallet init')")
	ErrRateMustBePositive = fmt.Errorf("rate must be a positive integer")

	svcLoadRates = config.LoadRates
	svcSaveRates = config.SaveRates
)

type TestRateConfig struct {
	BaseCurrency string
	Rates        map[string]int64
}

func SetTestRateConfig(cfg TestRateConfig) {
	svcLoadRates = func() (config.RateConfig, error) {
		rates := make(map[string]int64)
		for k, v := range cfg.Rates {
			rates[k] = v
		}
		return config.RateConfig{BaseCurrency: cfg.BaseCurrency, Rates: rates}, nil
	}
	svcSaveRates = func(rc config.RateConfig) error {
		cfg.BaseCurrency = rc.BaseCurrency
		cfg.Rates = rc.Rates
		return nil
	}
}

func ResetTestRateConfig() {
	svcLoadRates = config.LoadRates
	svcSaveRates = config.SaveRates
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

func (s *Service) loadRateConfig() (config.RateConfig, error) {
	cfg, err := svcLoadRates()
	if err != nil {
		if err.Error() == "rate configuration not found" {
			s.logger.Warn("loadRateConfig not found", slog.String("error", err.Error()))
			return config.RateConfig{}, ErrRateConfigMissing
		}
		s.logger.Error("loadRateConfig failed", slog.String("error", err.Error()))
		return config.RateConfig{}, fmt.Errorf("load rate config: %w", err)
	}
	return cfg, nil
}

func (s *Service) GetBaseCurrency() (string, error) {
	s.logger.Info("GetBaseCurrency called")
	cfg, err := s.loadRateConfig()
	if err != nil {
		return "", err
	}
	s.logger.Info("GetBaseCurrency completed", slog.String("base_currency", cfg.BaseCurrency))
	return cfg.BaseCurrency, nil
}

func (s *Service) GetRate(currency string) (int64, error) {
	s.logger.Info("GetRate called", slog.String("currency", currency))
	cfg, err := s.loadRateConfig()
	if err != nil {
		return 0, err
	}

	if currency == cfg.BaseCurrency {
		s.logger.Info("GetRate completed", slog.Int64("rate", 1))
		return 1, nil
	}

	rate, ok := cfg.Rates[currency]
	if !ok {
		s.logger.Warn("GetRate not found", slog.String("currency", currency))
		return 0, &RateNotFoundError{Currency: currency, Base: cfg.BaseCurrency}
	}

	if rate <= 0 {
		s.logger.Warn("GetRate rate not positive", slog.String("currency", currency), slog.Int64("rate", rate))
		return 0, fmt.Errorf("%w for %s", ErrRateMustBePositive, currency)
	}

	s.logger.Info("GetRate completed", slog.Int64("rate", rate))
	return rate, nil
}

func (s *Service) Convert(amount int64, fromCurrency string) (int64, error) {
	s.logger.Info("Convert called", slog.Int64("amount", amount), slog.String("from_currency", fromCurrency))
	cfg, err := s.loadRateConfig()
	if err != nil {
		return 0, err
	}

	if fromCurrency == cfg.BaseCurrency {
		s.logger.Info("Convert completed", slog.Int64("converted_amount", amount))
		return amount, nil
	}

	rate, ok := cfg.Rates[fromCurrency]
	if !ok {
		s.logger.Warn("Convert rate not found", slog.String("currency", fromCurrency))
		return 0, &RateNotFoundError{Currency: fromCurrency, Base: cfg.BaseCurrency}
	}

	if rate <= 0 {
		s.logger.Warn("Convert rate not positive", slog.String("currency", fromCurrency), slog.Int64("rate", rate))
		return 0, fmt.Errorf("%w for %s", ErrRateMustBePositive, fromCurrency)
	}

	result := int64(math.Round(float64(amount) * float64(rate)))
	s.logger.Info("Convert completed", slog.Int64("converted_amount", result))
	return result, nil
}

func (s *Service) ListRates() (string, map[string]int64, error) {
	s.logger.Info("ListRates called")
	cfg, err := s.loadRateConfig()
	if err != nil {
		return "", nil, err
	}
	s.logger.Info("ListRates completed", slog.String("base_currency", cfg.BaseCurrency), slog.Int("count", len(cfg.Rates)))
	return cfg.BaseCurrency, cfg.Rates, nil
}

func (s *Service) AddRate(currency string, rate int64) error {
	s.logger.Info("AddRate called", slog.String("currency", currency), slog.Int64("rate", rate))
	if rate <= 0 {
		s.logger.Warn("AddRate validation failed", slog.Int64("rate", rate))
		return fmt.Errorf("%w: %d", ErrRateMustBePositive, rate)
	}

	cfg, err := s.loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		s.logger.Warn("AddRate base currency rejected", slog.String("currency", currency))
		return fmt.Errorf("cannot add a rate for the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; exists {
		s.logger.Warn("AddRate already exists", slog.String("currency", currency))
		return fmt.Errorf("rate for %s already exists (use 'wallet rate set %s <rate>' to update)", currency, currency)
	}

	cfg.Rates[currency] = rate
	err = svcSaveRates(cfg)
	if err != nil {
		s.logger.Error("AddRate failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("AddRate completed", slog.String("currency", currency), slog.Int64("rate", rate))
	return nil
}

func (s *Service) SetRate(currency string, rate int64) error {
	s.logger.Info("SetRate called", slog.String("currency", currency), slog.Int64("rate", rate))
	if rate <= 0 {
		s.logger.Warn("SetRate validation failed", slog.Int64("rate", rate))
		return fmt.Errorf("%w: %d", ErrRateMustBePositive, rate)
	}

	cfg, err := s.loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		s.logger.Warn("SetRate base currency rejected", slog.String("currency", currency))
		return fmt.Errorf("cannot set a rate for the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; !exists {
		s.logger.Warn("SetRate not found", slog.String("currency", currency))
		return fmt.Errorf("no existing rate for %s (use 'wallet rate add %s <rate>' to add)", currency, currency)
	}

	cfg.Rates[currency] = rate
	err = svcSaveRates(cfg)
	if err != nil {
		s.logger.Error("SetRate failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("SetRate completed", slog.String("currency", currency), slog.Int64("rate", rate))
	return nil
}

func (s *Service) RemoveRate(currency string) error {
	s.logger.Info("RemoveRate called", slog.String("currency", currency))
	cfg, err := s.loadRateConfig()
	if err != nil {
		return err
	}

	if currency == cfg.BaseCurrency {
		s.logger.Warn("RemoveRate base currency rejected", slog.String("currency", currency))
		return fmt.Errorf("cannot remove the base currency '%s'", cfg.BaseCurrency)
	}

	if _, exists := cfg.Rates[currency]; !exists {
		s.logger.Warn("RemoveRate not found", slog.String("currency", currency))
		return fmt.Errorf("no configured rate for %s", currency)
	}

	delete(cfg.Rates, currency)
	err = svcSaveRates(cfg)
	if err != nil {
		s.logger.Error("RemoveRate failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("RemoveRate completed", slog.String("currency", currency))
	return nil
}
