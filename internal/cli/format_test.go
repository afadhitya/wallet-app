package cli

import (
	"strings"
	"testing"
)

func TestFormatAmount_Negative(t *testing.T) {
	result := formatAmount(-50000)
	if !strings.HasPrefix(result, "-") {
		t.Errorf("expected negative amount to start with '-', got %q", result)
	}
	if !strings.Contains(result, "50.000") {
		t.Errorf("expected formatted number '50.000' in result, got %q", result)
	}
}

func TestFormatAmount_Zero(t *testing.T) {
	result := formatAmount(0)
	expected := "Rp 0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatAmount_LargePositive(t *testing.T) {
	result := formatAmount(123456789)
	if !strings.Contains(result, "123.456.789") {
		t.Errorf("expected formatted number with dots, got %q", result)
	}
}

func TestFormatAmountWithCurrency_IDR(t *testing.T) {
	result := formatAmountWithCurrency(50000, "IDR")
	expected := "Rp 50.000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatAmountWithCurrency_USD(t *testing.T) {
	result := formatAmountWithCurrency(1000, "USD")
	expected := "$1.000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatAmountWithCurrency_EUR(t *testing.T) {
	result := formatAmountWithCurrency(500, "EUR")
	if !strings.HasPrefix(result, "\u20ac") {
		t.Errorf("expected euro symbol prefix, got %q", result)
	}
}

func TestFormatAmountWithCurrency_JPY(t *testing.T) {
	result := formatAmountWithCurrency(10000, "JPY")
	if !strings.HasPrefix(result, "\u00a5") {
		t.Errorf("expected yen symbol prefix, got %q", result)
	}
}

func TestFormatAmountWithCurrency_Negative(t *testing.T) {
	result := formatAmountWithCurrency(-50000, "IDR")
	expected := "-Rp 50.000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatAmountWithCurrency_Zero(t *testing.T) {
	result := formatAmountWithCurrency(0, "USD")
	expected := "$0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatAmountWithCurrency_Unknown(t *testing.T) {
	result := formatAmountWithCurrency(1000, "XYZ")
	if !strings.HasPrefix(result, "XYZ ") {
		t.Errorf("expected 'XYZ ' prefix for unknown currency, got %q", result)
	}
}
