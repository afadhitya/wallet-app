package shared

import (
	"fmt"
	"time"
)

func ParseDate(input string) (string, error) {
	if input == "" {
		return time.Now().Format("2006-01-02"), nil
	}

	switch input {
	case "today":
		return time.Now().Format("2006-01-02"), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "tomorrow":
		return time.Now().AddDate(0, 0, 1).Format("2006-01-02"), nil
	}

	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("02/01/2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("02 Jan 2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("2 Jan 2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", input)
}

func ParseMonth(input string) (string, string, error) {
	now := time.Now()

	months := map[string]time.Month{
		"january": 1, "february": 2, "march": 3, "april": 4,
		"may": 5, "june": 6, "july": 7, "august": 8,
		"september": 9, "october": 10, "november": 11, "december": 12,
		"jan": 1, "feb": 2, "mar": 3, "apr": 4,
		"jun": 6, "jul": 7, "aug": 8,
		"sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}

	month, ok := months[input]
	if !ok {
		if t, err := time.Parse("2006-01", input); err == nil {
			month = t.Month()
			now = time.Date(t.Year(), month, 1, 0, 0, 0, 0, time.UTC)
		} else if t, err := time.Parse("01/2006", input); err == nil {
			month = t.Month()
			now = time.Date(t.Year(), month, 1, 0, 0, 0, 0, time.UTC)
		} else {
			return "", "", fmt.Errorf("invalid month: %s", input)
		}
	}

	firstDay := time.Date(now.Year(), month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), nil
}
