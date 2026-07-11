package shared

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/testdb"
	"github.com/afadhitya/wallet-app/pkg/config"
)

func setupQuerier(t *testing.T) gen.Querier {
	t.Helper()
	return gen.New(testdb.Open(t))
}

func TestNotFoundErrorUnwrap(t *testing.T) {
	e := &NotFoundError{Entity: "account", Name: "test"}
	if !errors.Is(e, ErrNotFound) {
		t.Error("expected NotFoundError to unwrap to ErrNotFound")
	}
	if errors.Unwrap(e) != ErrNotFound {
		t.Error("expected Unwrap to return ErrNotFound")
	}
}

func TestValidationErrorError(t *testing.T) {
	e := &ValidationError{Field: "name", Message: "cannot be empty"}
	if e.Error() != "name: cannot be empty" {
		t.Errorf("expected 'name: cannot be empty', got '%s'", e.Error())
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-07-01", "2026-07-01"},
		{"01/07/2026", "2026-07-01"},
		{"01 Jul 2026", "2026-07-01"},
		{"1 Jul 2026", "2026-07-01"},
		{"", ""},
	}

	for _, tc := range tests {
		result, err := ParseDate(tc.input)
		if tc.expected == "" {
			if err != nil {
				continue
			}
			if result == "" {
				t.Error("expected today's date for empty input, got empty")
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDate(%q): %v", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("ParseDate(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestParseDateToday(t *testing.T) {
	result, err := ParseDate("today")
	if err != nil {
		t.Fatalf("ParseDate(today): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'today'")
	}
}

func TestParseDateYesterday(t *testing.T) {
	result, err := ParseDate("yesterday")
	if err != nil {
		t.Fatalf("ParseDate(yesterday): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'yesterday'")
	}
}

func TestParseDateInvalid(t *testing.T) {
	_, err := ParseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestParseMonth(t *testing.T) {
	_, _, err := ParseMonth("july")
	if err != nil {
		t.Fatalf("ParseMonth(july): %v", err)
	}

	_, _, err = ParseMonth("jan")
	if err != nil {
		t.Fatalf("ParseMonth(jan): %v", err)
	}

	_, _, err = ParseMonth("2026-07")
	if err != nil {
		t.Fatalf("ParseMonth(2026-07): %v", err)
	}
}

func TestParseMonthInvalid(t *testing.T) {
	_, _, err := ParseMonth("not-a-month")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestParseMonthSlashFormat(t *testing.T) {
	from, to, err := ParseMonth("07/2026")
	if err != nil {
		t.Fatalf("ParseMonth(07/2026): %v", err)
	}
	if from != "2026-07-01" {
		t.Errorf("expected from '2026-07-01', got '%s'", from)
	}
	if to != "2026-07-31" {
		t.Errorf("expected to '2026-07-31', got '%s'", to)
	}
}

func TestParseDateTomorrow(t *testing.T) {
	result, err := ParseDate("tomorrow")
	if err != nil {
		t.Fatalf("ParseDate(tomorrow): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty date for 'tomorrow'")
	}
}

func TestResolveAccountByID(t *testing.T) {
	q := setupQuerier(t)

	account, err := q.CreateAccount(context.Background(), gen.CreateAccountParams{
		Name: "BCA", Type: "checking", Currency: "IDR",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	resolved, err := ResolveAccount(q, "1")
	if err != nil {
		t.Fatalf("ResolveAccount by ID: %v", err)
	}
	if resolved.ID != account.ID {
		t.Errorf("expected ID %d, got %d", account.ID, resolved.ID)
	}
}

func TestResolveAccountByName(t *testing.T) {
	q := setupQuerier(t)

	_, err := q.CreateAccount(context.Background(), gen.CreateAccountParams{
		Name: "BCA", Type: "checking", Currency: "IDR",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	resolved, err := ResolveAccount(q, "BCA")
	if err != nil {
		t.Fatalf("ResolveAccount by name: %v", err)
	}
	if resolved.Name != "BCA" {
		t.Errorf("expected name 'BCA', got '%s'", resolved.Name)
	}
}

func TestResolveAccountNotFound(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveAccount(q, "Ghost")
	if err == nil {
		t.Fatal("expected error for unknown account")
	}
}

func TestResolveAccountByIDNotFoundFallsToName(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveAccount(q, "9999")
	if err == nil {
		t.Fatal("expected error for non-existent identifier")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveCategoryWithSuggestions(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveCategory(q, "Restauran")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestResolveCategoryByID(t *testing.T) {
	q := setupQuerier(t)

	cat, err := q.CreateCategory(context.Background(), gen.CreateCategoryParams{Name: "CustomCat"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	resolved, err := ResolveCategory(q, fmt.Sprintf("%d", cat.ID))
	if err != nil {
		t.Fatalf("ResolveCategory by ID: %v", err)
	}
	if resolved.ID != cat.ID {
		t.Errorf("expected ID %d, got %d", cat.ID, resolved.ID)
	}
}

func TestResolveCategoryByIDNotFoundFallsToName(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveCategory(q, "9999")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestResolveTagByID(t *testing.T) {
	q := setupQuerier(t)

	tag, err := q.CreateTag(context.Background(), "test-tag")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	resolved, err := ResolveTag(q, fmt.Sprintf("%d", tag.ID))
	if err != nil {
		t.Fatalf("ResolveTag by ID: %v", err)
	}
	if resolved.ID != tag.ID {
		t.Errorf("expected ID %d, got %d", tag.ID, resolved.ID)
	}
}

func TestResolveTagByIDNotFoundFallsToName(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveTag(q, "9999")
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestResolveTagByNameNotFound(t *testing.T) {
	q := setupQuerier(t)

	_, err := ResolveTag(q, "nonexistent-tag")
	if err == nil {
		t.Fatal("expected error for non-existent tag name")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetBaseCurrency(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	base, err := GetBaseCurrency()
	if err != nil {
		t.Fatalf("GetBaseCurrency: %v", err)
	}
	if base != "IDR" {
		t.Errorf("expected 'IDR', got '%s'", base)
	}
}

func TestGetRate(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	rate, err := GetRate("USD")
	if err != nil {
		t.Fatalf("GetRate USD: %v", err)
	}
	if rate != 15800 {
		t.Errorf("expected 15800, got %d", rate)
	}
}

func TestGetRateBaseCurrencyIsOne(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	rate, err := GetRate("IDR")
	if err != nil {
		t.Fatalf("GetRate base: %v", err)
	}
	if rate != 1 {
		t.Errorf("expected 1, got %d", rate)
	}
}

func TestGetRateMissing(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	_, err := GetRate("KRW")
	if err == nil {
		t.Fatal("expected error for missing rate")
	}
	var rnf *RateNotFoundError
	if !errors.As(err, &rnf) {
		t.Errorf("expected RateNotFoundError, got %T: %v", err, err)
	}
}

func TestConvert(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	converted, err := Convert(10, "USD")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	expected := int64(10 * 15800)
	if converted != expected {
		t.Errorf("expected %d, got %d", expected, converted)
	}
}

func TestConvertBaseCurrencyIsIdentity(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	converted, err := Convert(50000, "IDR")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if converted != 50000 {
		t.Errorf("expected 50000, got %d", converted)
	}
}

func TestConvertMissingRate(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	_, err := Convert(5000, "KRW")
	if err == nil {
		t.Fatal("expected error for missing rate")
	}
}

func TestConvertRounding(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"JPY": 105},
	})
	defer ResetTestRateConfig()

	converted, err := Convert(100, "JPY")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if converted != 10500 {
		t.Errorf("expected 10500, got %d", converted)
	}
}

func TestListRates(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})
	defer ResetTestRateConfig()

	base, rates, err := ListRates()
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
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := AddRate("KRW", 12)
	if err != nil {
		t.Fatalf("AddRate: %v", err)
	}

	rate, err := GetRate("KRW")
	if err != nil {
		t.Fatalf("GetRate after add: %v", err)
	}
	if rate != 12 {
		t.Errorf("expected 12, got %d", rate)
	}
}

func TestAddRateNonPositive(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := AddRate("KRW", -5)
	if err == nil {
		t.Fatal("expected error for non-positive rate")
	}
	if !errors.Is(err, ErrRateMustBePositive) {
		t.Errorf("expected ErrRateMustBePositive, got: %v", err)
	}
}

func TestAddRateAlreadyExists(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := AddRate("USD", 16000)
	if err == nil {
		t.Fatal("expected error for existing rate")
	}
}

func TestAddRateBaseCurrency(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := AddRate("IDR", 1)
	if err == nil {
		t.Fatal("expected error for adding rate for base currency")
	}
}

func TestSetRate(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := SetRate("USD", 16000)
	if err != nil {
		t.Fatalf("SetRate: %v", err)
	}

	rate, err := GetRate("USD")
	if err != nil {
		t.Fatalf("GetRate after set: %v", err)
	}
	if rate != 16000 {
		t.Errorf("expected 16000, got %d", rate)
	}
}

func TestSetRateNonExistent(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := SetRate("KRW", 12)
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
}

func TestSetRateNonPositive(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := SetRate("USD", 0)
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
	if !errors.Is(err, ErrRateMustBePositive) {
		t.Errorf("expected ErrRateMustBePositive, got: %v", err)
	}
}

func TestSetRateBaseCurrency(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := SetRate("IDR", 1)
	if err == nil {
		t.Fatal("expected error for setting base currency rate")
	}
}

func TestRemoveRate(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates: map[string]int64{
			"USD": 15800,
			"EUR": 17200,
		},
	})
	defer ResetTestRateConfig()

	err := RemoveRate("EUR")
	if err != nil {
		t.Fatalf("RemoveRate: %v", err)
	}

	_, err = GetRate("EUR")
	if err == nil {
		t.Fatal("expected error after removal")
	}
}

func TestRemoveRateNonExistent(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := RemoveRate("KRW")
	if err == nil {
		t.Fatal("expected error for non-existent rate")
	}
}

func TestRemoveRateBaseCurrency(t *testing.T) {
	SetTestRateConfig(TestRateConfig{
		BaseCurrency: "IDR",
		Rates:        map[string]int64{"USD": 15800},
	})
	defer ResetTestRateConfig()

	err := RemoveRate("IDR")
	if err == nil {
		t.Fatal("expected error for removing base currency")
	}
}

func TestRateConfigMissingError(t *testing.T) {
	origLoad := LoadRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, errors.New("rate configuration not found")
	}
	defer func() { LoadRates = origLoad }()

	_, err := GetBaseCurrency()
	if err == nil {
		t.Fatal("expected error for missing rate config")
	}
	if !errors.Is(err, ErrRateConfigMissing) {
		t.Errorf("expected ErrRateConfigMissing, got: %v", err)
	}
}

func TestRateConfigLoadError(t *testing.T) {
	origLoad := LoadRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, errors.New("read rates file: permission denied")
	}
	defer func() { LoadRates = origLoad }()

	_, err := GetBaseCurrency()
	if err == nil {
		t.Fatal("expected error for load failure")
	}
}

func TestGetRateNonPositiveInConfig(t *testing.T) {
	origLoad := LoadRates
	origSave := SaveRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": 0},
		}, nil
	}
	SaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		LoadRates = origLoad
		SaveRates = origSave
	}()

	_, err := GetRate("USD")
	if err == nil {
		t.Fatal("expected error for non-positive configured rate")
	}
	if !errors.Is(err, ErrRateMustBePositive) {
		t.Errorf("expected ErrRateMustBePositive, got: %v", err)
	}
}

func TestConvertNonPositiveRate(t *testing.T) {
	origLoad := LoadRates
	origSave := SaveRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": -1},
		}, nil
	}
	SaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		LoadRates = origLoad
		SaveRates = origSave
	}()

	_, err := Convert(100, "USD")
	if err == nil {
		t.Fatal("expected error for non-positive rate in convert")
	}
}

func TestConvertMissingRateInConvert(t *testing.T) {
	origLoad := LoadRates
	origSave := SaveRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{
			BaseCurrency: "IDR",
			Rates:        map[string]int64{"USD": 15800},
		}, nil
	}
	SaveRates = func(config.RateConfig) error { return nil }
	defer func() {
		LoadRates = origLoad
		SaveRates = origSave
	}()

	_, err := Convert(100, "KRW")
	if err == nil {
		t.Fatal("expected error for missing rate in convert")
	}
}

func TestListRatesWithLoadError(t *testing.T) {
	origLoad := LoadRates
	LoadRates = func() (config.RateConfig, error) {
		return config.RateConfig{}, ErrRateConfigMissing
	}
	defer func() { LoadRates = origLoad }()

	_, _, err := ListRates()
	if err == nil {
		t.Fatal("expected error for load failure")
	}
}
