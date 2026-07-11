package plannedpayment

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func ComputeInitialDueDate(startDate string, recurrence string, dueDay int) string {
	if recurrence == "none" {
		return startDate
	}
	t, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return startDate
	}
	if dueDay <= 0 {
		return t.Format("2006-01-02")
	}
	switch recurrence {
	case "weekly":
		weekday := time.Weekday(dueDay)
		if weekday < time.Monday || weekday > time.Sunday {
			return t.Format("2006-01-02")
		}
		diff := (int(weekday) - int(t.Weekday()) + 7) % 7
		return t.AddDate(0, 0, diff).Format("2006-01-02")
	case "monthly", "custom":
		return setDayInMonth(t.Year(), int(t.Month()), dueDay)
	case "yearly":
		return setDayInMonth(t.Year(), int(t.Month()), dueDay)
	default:
		return t.Format("2006-01-02")
	}
}

func setDayInMonth(year, month, day int) string {
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func CalcNextDue(currentDue time.Time, recurrence string, recurrenceRule sql.NullString) (time.Time, error) {
	switch recurrence {
	case "daily":
		return currentDue.AddDate(0, 0, 1), nil
	case "weekly":
		return currentDue.AddDate(0, 0, 7), nil
	case "monthly":
		year := currentDue.Year()
		month := currentDue.Month() + 1
		if month > 12 {
			month = 1
			year++
		}
		day := currentDue.Day()
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
	case "yearly":
		return currentDue.AddDate(1, 0, 0), nil
	case "custom":
		if !recurrenceRule.Valid {
			return time.Time{}, &shared.ValidationError{Field: "recurrence_rule", Message: "custom recurrence rule is required"}
		}
		return calcNextDueFromRRULE(currentDue, recurrenceRule.String)
	case "none":
		return currentDue, nil
	default:
		return time.Time{}, &shared.ValidationError{Field: "recurrence", Message: fmt.Sprintf("unknown recurrence: %s", recurrence)}
	}
}

var bydayLookup = map[string]time.Weekday{
	"MO": time.Monday,
	"TU": time.Tuesday,
	"WE": time.Wednesday,
	"TH": time.Thursday,
	"FR": time.Friday,
	"SA": time.Saturday,
	"SU": time.Sunday,
}

func calcNextDueFromRRULE(currentDue time.Time, rrule string) (time.Time, error) {
	rule := strings.ToUpper(rrule)
	if !strings.HasPrefix(rule, "FREQ=") {
		return time.Time{}, &shared.ValidationError{Field: "rrule", Message: "RRULE must start with FREQ="}
	}

	parts := strings.Split(rule, ";")
	freq := strings.TrimPrefix(parts[0], "FREQ=")
	var byMonthDay int
	bydayDays := make(map[time.Weekday]bool)

	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "BYMONTHDAY=") {
			_, _ = fmt.Sscanf(part, "BYMONTHDAY=%d", &byMonthDay)
		}
		if strings.HasPrefix(part, "BYDAY=") {
			dayStr := strings.TrimPrefix(part, "BYDAY=")
			for _, d := range strings.Split(dayStr, ",") {
				if wd, ok := bydayLookup[strings.TrimSpace(d)]; ok {
					bydayDays[wd] = true
				}
			}
		}
	}

	switch freq {
	case "DAILY":
		return currentDue.AddDate(0, 0, 1), nil
	case "WEEKLY":
		if len(bydayDays) > 0 {
			start := currentDue.AddDate(0, 0, 1)
			for i := 0; i < 7; i++ {
				if bydayDays[start.Weekday()] {
					return start, nil
				}
				start = start.AddDate(0, 0, 1)
			}
		}
		return currentDue.AddDate(0, 0, 7), nil
	case "MONTHLY":
		year := currentDue.Year()
		month := currentDue.Month() + 1
		if month > 12 {
			month = 1
			year++
		}
		day := currentDue.Day()
		if byMonthDay > 0 {
			day = byMonthDay
		}
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
	case "YEARLY":
		return currentDue.AddDate(1, 0, 0), nil
	default:
		return time.Time{}, &shared.ValidationError{Field: "rrule", Message: fmt.Sprintf("unsupported RRULE frequency: %s", freq)}
	}
}

func ValidateRRULE(rrule string) error {
	rule := strings.ToUpper(strings.TrimSpace(rrule))
	if rule == "" {
		return fmt.Errorf("recurrence rule cannot be empty")
	}
	if !strings.HasPrefix(rule, "FREQ=") {
		return fmt.Errorf("RRULE must start with FREQ=")
	}
	parts := strings.Split(rule, ";")
	freq := strings.TrimPrefix(parts[0], "FREQ=")
	validFreqs := map[string]bool{"DAILY": true, "WEEKLY": true, "MONTHLY": true, "YEARLY": true}
	if !validFreqs[freq] {
		return fmt.Errorf("unsupported RRULE frequency: %s (use DAILY, WEEKLY, MONTHLY, or YEARLY)", freq)
	}
	return nil
}
