package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func formatAmount(amount int64) string {
	if amount < 0 {
		return fmt.Sprintf("-Rp %s", formatNum(-amount))
	}
	return fmt.Sprintf("Rp %s", formatNum(amount))
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

func formatDate(date string) string {
	if date == "" {
		return "-"
	}
	return date
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func joinTagNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
}
