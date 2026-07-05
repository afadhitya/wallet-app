package cli

import (
	"fmt"
	"strconv"
)

var currencySymbols = map[string]string{
	"IDR": "Rp ",
	"USD": "$",
	"EUR": "\u20ac",
	"JPY": "\u00a5",
	"SGD": "S$",
	"GBP": "\u00a3",
	"AUD": "A$",
	"CNY": "\u00a5",
	"KRW": "\u20a9",
	"MYR": "RM",
	"THB": "\u0e3f",
	"PHP": "\u20b1",
	"INR": "\u20b9",
	"VND": "\u20ab",
	"HKD": "HK$",
	"CHF": "CHF ",
	"CAD": "C$",
	"NZD": "NZ$",
	"SAR": "SAR ",
	"AED": "AED ",
}

func formatAmount(amount int64) string {
	if amount < 0 {
		return fmt.Sprintf("-Rp %s", formatNum(-amount))
	}
	return fmt.Sprintf("Rp %s", formatNum(amount))
}

func formatAmountWithCurrency(amount int64, currency string) string {
	symbol, ok := currencySymbols[currency]
	if !ok {
		symbol = currency + " "
	}
	if amount < 0 {
		return fmt.Sprintf("-%s%s", symbol, formatNum(-amount))
	}
	return fmt.Sprintf("%s%s", symbol, formatNum(amount))
}

func formatNum(n int64) string {
	s := strconv.FormatInt(n, 10)
	parts := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			parts = append(parts, '.')
		}
		parts = append(parts, byte(c))
	}
	return string(parts)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
