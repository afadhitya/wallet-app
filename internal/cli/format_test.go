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
