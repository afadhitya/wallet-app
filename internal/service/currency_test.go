package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/afadhitya/wallet-app/internal/testdb"
	"github.com/afadhitya/wallet-app/pkg/config"
)

func setupServiceWithRates(t *testing.T, base string, rates map[string]int64) *Service {
	t.Helper()

	origLoad := svcLoadRates
	origSave := svcSaveRates

	stored := config.RateConfig{
		BaseCurrency: base,
		Rates:        make(map[string]int64),
	}
	for k, v := range rates {
		stored.Rates[k] = v
	}

	svcLoadRates = func() (config.RateConfig, error) {
		rc := config.RateConfig{
			BaseCurrency: stored.BaseCurrency,
			Rates:        make(map[string]int64),
		}
		for k, v := range stored.Rates {
			rc.Rates[k] = v
		}
		return rc, nil
	}

	svcSaveRates = func(cfg config.RateConfig) error {
		stored = cfg
		return nil
	}

	t.Cleanup(func() {
		svcLoadRates = origLoad
		svcSaveRates = origSave
	})

	return New(testdb.Open(t))
}

func TestGetBaseCurrency(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	base, err := svc.GetBaseCurrency()
	if err != nil {
		t.Fatalf("GetBaseCurrency: %v", err)
	}
	if base != "IDR" {
		t.Errorf("expected 'IDR', got '%s'", base)
	}
}

func TestGetRate(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	rate, err := svc.GetRate("USD")
	if err != nil {
		t.Fatalf("GetRate USD: %v", err)
	}
	if rate != 15800 {
		t.Errorf("expected 15800, got %d", rate)
	}
}

func TestGetRateBaseCurrencyIsOne(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	rate, err := svc.GetRate("IDR")
	if err != nil {
		t.Fatalf("GetRate base: %v", err)
	}
	if rate != 1 {
		t.Errorf("expected 1, got %d", rate)
	}
}

func TestGetRateMissing(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	_, err := svc.GetRate("KRW")
	if err == nil {
		t.Fatal("expected error for missing rate")
	}
	var rnf *RateNotFoundError
	if !errors.As(err, &rnf) {
		t.Errorf("expected RateNotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "wallet rate add KRW") {
		t.Errorf("expected actionable error, got: %v", err)
	}
}

func TestConvert(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	converted, err := svc.Convert(10, "USD")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	expected := int64(10 * 15800)
	if converted != expected {
		t.Errorf("expected %d, got %d", expected, converted)
	}
}

func TestConvertBaseCurrencyIsIdentity(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	converted, err := svc.Convert(50000, "IDR")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if converted != 50000 {
		t.Errorf("expected 50000, got %d", converted)
	}
}

func TestConvertMissingRate(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	_, err := svc.Convert(5000, "KRW")
	if err == nil {
		t.Fatal("expected error for missing rate")
	}
	if !strings.Contains(err.Error(), "wallet rate add KRW") {
		t.Errorf("expected actionable error, got: %v", err)
	}
}

func TestConvertRounding(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"JPY": 105})

	converted, err := svc.Convert(100, "JPY")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if converted != 10500 {
		t.Errorf("expected 10500, got %d", converted)
	}
}

func TestListRates(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{
		"USD": 15800,
		"EUR": 17200,
	})

	base, rates, err := svc.ListRates()
	if err != nil {
		t.Fatalf("ListRates: %v", err)
	}
	if base != "IDR" {
		t.Errorf("expected base 'IDR', got '%s'", base)
	}
	if rates["USD"] != 15800 {
		t.Errorf("expected USD 15800, got %d", rates["USD"])
	}
	if rates["EUR"] != 17200 {
		t.Errorf("expected EUR 17200, got %d", rates["EUR"])
	}
}

func TestAddRate(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.AddRate("KRW", 12)
	if err != nil {
		t.Fatalf("AddRate: %v", err)
	}

	rate, err := svc.GetRate("KRW")
	if err != nil {
		t.Fatalf("GetRate after add: %v", err)
	}
	if rate != 12 {
		t.Errorf("expected 12, got %d", rate)
	}
}

func TestAddRateNonPositive(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.AddRate("KRW", -5)
	if err == nil {
		t.Fatal("expected error for non-positive rate")
	}
	if !errors.Is(err, ErrRateMustBePositive) {
		t.Errorf("expected ErrRateMustBePositive, got: %v", err)
	}
}

func TestAddRateAlreadyExists(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.AddRate("USD", 16000)
	if err == nil {
		t.Fatal("expected error for existing rate")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestAddRateBaseCurrency(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.AddRate("IDR", 1)
	if err == nil {
		t.Fatal("expected error for adding rate for base currency")
	}
	if !strings.Contains(err.Error(), "cannot add a rate for the base currency") {
		t.Errorf("expected base currency error, got: %v", err)
	}
}

func TestSetRate(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.SetRate("USD", 16000)
	if err != nil {
		t.Fatalf("SetRate: %v", err)
	}

	rate, err := svc.GetRate("USD")
	if err != nil {
		t.Fatalf("GetRate after set: %v", err)
	}
	if rate != 16000 {
		t.Errorf("expected 16000, got %d", rate)
	}
}

func TestSetRateNonExistent(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.SetRate("KRW", 12)
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
	if !strings.Contains(err.Error(), "no existing rate") {
		t.Errorf("expected 'no existing rate' error, got: %v", err)
	}
}

func TestSetRateNonPositive(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.SetRate("USD", 0)
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
	if !errors.Is(err, ErrRateMustBePositive) {
		t.Errorf("expected ErrRateMustBePositive, got: %v", err)
	}
}

func TestSetRateBaseCurrency(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.SetRate("IDR", 1)
	if err == nil {
		t.Fatal("expected error for setting base currency rate")
	}
	if !strings.Contains(err.Error(), "cannot set a rate for the base currency") {
		t.Errorf("expected base currency error, got: %v", err)
	}
}

func TestRemoveRate(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{
		"USD": 15800,
		"EUR": 17200,
	})

	err := svc.RemoveRate("EUR")
	if err != nil {
		t.Fatalf("RemoveRate: %v", err)
	}

	_, err = svc.GetRate("EUR")
	if err == nil {
		t.Fatal("expected error after removal")
	}
}

func TestRemoveRateNonExistent(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.RemoveRate("KRW")
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
	if !strings.Contains(err.Error(), "no configured rate") {
		t.Errorf("expected 'no configured rate' error, got: %v", err)
	}
}

func TestRemoveRateBaseCurrency(t *testing.T) {
	svc := setupServiceWithRates(t, "IDR", map[string]int64{"USD": 15800})

	err := svc.RemoveRate("IDR")
	if err == nil {
		t.Fatal("expected error for removing base currency")
	}
	if !strings.Contains(err.Error(), "cannot remove the base currency") {
		t.Errorf("expected base currency error, got: %v", err)
	}
}

func TestRateConfigMissingError(t *testing.T) {
	origLoad := svcLoadRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, errors.New("rate configuration not found")
	}
	defer func() { svcLoadRates = origLoad }()

	svc := New(testdb.Open(t))
	_, err := svc.GetBaseCurrency()
	if err == nil {
		t.Fatal("expected error for missing rate config")
	}
	if !errors.Is(err, ErrRateConfigMissing) {
		t.Errorf("expected ErrRateConfigMissing, got: %v", err)
	}
}

func TestRateConfigLoadError(t *testing.T) {
	origLoad := svcLoadRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, errors.New("read rates file: permission denied")
	}
	defer func() { svcLoadRates = origLoad }()

	svc := New(testdb.Open(t))
	_, err := svc.GetBaseCurrency()
	if err == nil {
		t.Fatal("expected error for load failure")
	}
	if !strings.Contains(err.Error(), "load rate config:") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestGetRateNonPositiveInConfig(t *testing.T) {
	origLoad := svcLoadRates
	origSave := svcSaveRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": 0},
		}, nil
	}
	svcSaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		svcLoadRates = origLoad
		svcSaveRates = origSave
	}()

	svc := New(testdb.Open(t))
	_, err := svc.GetRate("USD")
	if err == nil {
		t.Fatal("expected error for non-positive configured rate")
	}
	if !strings.Contains(err.Error(), "must be a positive integer") {
		t.Errorf("expected positive integer error, got: %v", err)
	}
}

func TestConvertNonPositiveRate(t *testing.T) {
	origLoad := svcLoadRates
	origSave := svcSaveRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": -1},
		}, nil
	}
	svcSaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		svcLoadRates = origLoad
		svcSaveRates = origSave
	}()

	svc := New(testdb.Open(t))
	_, err := svc.Convert(100, "USD")
	if err == nil {
		t.Fatal("expected error for non-positive rate in convert")
	}
}

func TestConvertMissingRateInConvert(t *testing.T) {
	origLoad := svcLoadRates
	origSave := svcSaveRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": 15800},
		}, nil
	}
	svcSaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		svcLoadRates = origLoad
		svcSaveRates = origSave
	}()

	svc := New(testdb.Open(t))
	_, err := svc.Convert(100, "KRW")
	if err == nil {
		t.Fatal("expected error for missing rate in convert")
	}
}

func TestListRatesWithLoadError(t *testing.T) {
	origLoad := svcLoadRates
	svcLoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, ErrRateConfigMissing
	}
	defer func() { svcLoadRates = origLoad }()

	svc := New(testdb.Open(t))
	_, _, err := svc.ListRates()
	if err == nil {
		t.Fatal("expected error for load failure")
	}
}
