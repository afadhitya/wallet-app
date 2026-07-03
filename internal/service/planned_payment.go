package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type CreatePlannedPaymentParams struct {
	Name           string
	Amount         int64
	Currency       string
	Type           string
	Account        string
	Category       string
	Recurrence     string
	RecurrenceRule string
	StartDate      string
	DueDay         int
}

type EditPlannedPaymentParams struct {
	Name           *string
	Amount         *int64
	Currency       *string
	Type           *string
	Account        *string
	Category       *string
	Recurrence     *string
	RecurrenceRule *string
	StartDate      *string
	DueDay         *int
}

type PayPlannedPaymentParams struct {
	ID     int64
	Date   string
	Amount int64
}

type PayPlannedPaymentResult struct {
	Transaction    *gen.Transaction
	PlannedPayment *gen.PlannedPayment
	NextDueDate    string
}

type ListDueFilter int

const (
	DueCurrentMonth ListDueFilter = iota
	DueCurrentWeek
	DueOverdue
	DueNextDays
)

type ListDueParams struct {
	Filter   ListDueFilter
	NextDays int
}

func (s *Service) CreatePlannedPayment(params CreatePlannedPaymentParams) (*gen.PlannedPayment, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if params.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if params.Currency == "" {
		params.Currency = "IDR"
	}
	if params.Type == "" {
		params.Type = "expense"
	}

	validRecurrences := map[string]bool{
		"none": true, "daily": true, "weekly": true,
		"monthly": true, "yearly": true, "custom": true,
	}
	if !validRecurrences[params.Recurrence] {
		return nil, &ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
	}

	if params.Recurrence == "custom" {
		if params.RecurrenceRule == "" {
			return nil, &ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
		}
		if err := validateRRULE(params.RecurrenceRule); err != nil {
			return nil, &ValidationError{Field: "rrule", Message: err.Error()}
		}
	}

	account, err := s.ResolveAccount(params.Account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	var categoryID sql.NullInt64
	if params.Category != "" {
		category, err := s.ResolveCategory(params.Category)
		if err != nil {
			return nil, fmt.Errorf("category: %w", err)
		}
		categoryID = sql.NullInt64{Int64: category.ID, Valid: true}
	}

	startDate, err := parseDate(params.StartDate)
	if err != nil {
		return nil, &ValidationError{Field: "start_date", Message: err.Error()}
	}

	var recurrenceRule sql.NullString
	if params.RecurrenceRule != "" {
		recurrenceRule = sql.NullString{String: params.RecurrenceRule, Valid: true}
	}

	nextDueDate := computeInitialDueDate(startDate, params.Recurrence, params.DueDay)
	var nextDueDateNull sql.NullString
	if nextDueDate != "" {
		nextDueDateNull = sql.NullString{String: nextDueDate, Valid: true}
	}

	pp, err := s.q.CreatePlannedPayment(s.ctx(), gen.CreatePlannedPaymentParams{
		AccountID:      account.ID,
		CategoryID:     categoryID,
		Type:           params.Type,
		Amount:         params.Amount,
		Currency:       params.Currency,
		Name:           params.Name,
		Recurrence:     params.Recurrence,
		RecurrenceRule: recurrenceRule,
		StartDate:      startDate,
		NextDueDate:    nextDueDateNull,
	})
	if err != nil {
		return nil, fmt.Errorf("create planned payment: %w", err)
	}

	return pp, nil
}

func (s *Service) GetPlannedPaymentByID(id int64) (*gen.PlannedPayment, error) {
	pp, err := s.q.GetPlannedPaymentByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "planned payment", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}
	return pp, nil
}

func (s *Service) GetAccountName(id int64) (string, error) {
	acc, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		return "", err
	}
	return acc.Name, nil
}

func (s *Service) GetCategoryName(id int64) string {
	cat, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		return ""
	}
	return cat.Name
}

func (s *Service) ListPlannedPayments(includePaused, includeArchived bool) ([]*gen.PlannedPayment, error) {
	if includeArchived {
		if includePaused {
			return s.q.ListAllPlannedPayments(s.ctx())
		}
		active, err := s.q.ListActivePlannedPayments(s.ctx())
		if err != nil {
			return nil, err
		}
		archived, err := s.q.ListArchivedPlannedPayments(s.ctx())
		if err != nil {
			return nil, err
		}
		result := make([]*gen.PlannedPayment, 0, len(active)+len(archived))
		result = append(result, active...)
		result = append(result, archived...)
		return result, nil
	}
	if includePaused {
		return s.q.ListPausedPlannedPayments(s.ctx())
	}
	return s.q.ListActivePlannedPayments(s.ctx())
}

func (s *Service) ListDuePlannedPayments(params ListDueParams) ([]*gen.PlannedPayment, int64, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	var dateFrom, dateTo string

	switch params.Filter {
	case DueOverdue:
		overdue, err := s.q.ListOverduePlannedPayments(s.ctx(), sql.NullString{String: today, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		var total int64
		for _, pp := range overdue {
			total += pp.Amount
		}
		return overdue, total, nil
	case DueCurrentWeek:
		weekday := now.Weekday()
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		monday := now.AddDate(0, 0, -daysToMonday)
		sunday := monday.AddDate(0, 0, 6)
		dateFrom = monday.Format("2006-01-02")
		dateTo = sunday.Format("2006-01-02")
	case DueNextDays:
		dateFrom = today
		dateTo = now.AddDate(0, 0, params.NextDays).Format("2006-01-02")
	default:
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, -1)
		dateFrom = firstDay.Format("2006-01-02")
		dateTo = lastDay.Format("2006-01-02")
	}

	due, err := s.q.ListDuePlannedPayments(s.ctx(), gen.ListDuePlannedPaymentsParams{
		DateFrom: sql.NullString{String: dateFrom, Valid: true},
		DateTo:   sql.NullString{String: dateTo, Valid: true},
	})
	if err != nil {
		return nil, 0, err
	}

	var total int64
	for _, pp := range due {
		total += pp.Amount
	}
	return due, total, nil
}

func (s *Service) PayPlannedPayment(params PayPlannedPaymentParams) (*PayPlannedPaymentResult, error) {
	pp, err := s.GetPlannedPaymentByID(params.ID)
	if err != nil {
		return nil, err
	}
	if pp.IsPaused != 0 {
		return nil, &ValidationError{Field: "state", Message: "cannot pay a paused planned payment"}
	}

	amount := pp.Amount
	if params.Amount > 0 {
		amount = params.Amount
	}

	date := ""
	if params.Date != "" {
		if parsed, err := parseDate(params.Date); err == nil {
			date = parsed
		} else {
			return nil, &ValidationError{Field: "date", Message: err.Error()}
		}
	} else {
		date = time.Now().Format("2006-01-02")
	}

	account, err := s.q.GetAccountByID(s.ctx(), pp.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	var description sql.NullString
	if pp.Name != "" {
		description = sql.NullString{String: pp.Name, Valid: true}
	}
	var plannedPaymentID sql.NullInt64
	plannedPaymentID = sql.NullInt64{Int64: pp.ID, Valid: true}

	txn, err := s.q.CreatePlannedTransaction(s.ctx(), gen.CreatePlannedTransactionParams{
		AccountID:        pp.AccountID,
		CategoryID:       pp.CategoryID,
		Type:             pp.Type,
		Amount:           amount,
		Currency:         pp.Currency,
		Description:      description,
		Notes:            sql.NullString{},
		TransferToID:     sql.NullInt64{},
		Date:             date,
		PlannedPaymentID: plannedPaymentID,
	})
	if err != nil {
		return nil, fmt.Errorf("create planned transaction: %w", err)
	}

	if err := s.recalculateBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	var nextDueDate string
	if pp.Recurrence == "none" {
		if err := s.q.ArchivePlannedPayment(s.ctx(), pp.ID); err != nil {
			return nil, fmt.Errorf("archive planned payment: %w", err)
		}
		pp.IsActive = 0
	} else {
		currentDue := ""
		if pp.NextDueDate.Valid {
			currentDue = pp.NextDueDate.String
		} else {
			currentDue = pp.StartDate
		}
		dueTime, err := parseDate(currentDue)
		if err != nil {
			return nil, fmt.Errorf("parse current due date: %w", err)
		}
		dueDate, _ := time.Parse("2006-01-02", dueTime)
		newDue, err := calcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
		if err != nil {
			return nil, err
		}
		nextDueDate = newDue.Format("2006-01-02")
		updated, err := s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
			ID:          pp.ID,
			NextDueDate: sql.NullString{String: nextDueDate, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("update next due date: %w", err)
		}
		pp = updated
	}

	return &PayPlannedPaymentResult{
		Transaction:    txn,
		PlannedPayment: pp,
		NextDueDate:    nextDueDate,
	}, nil
}

func (s *Service) SkipPlannedPayment(id int64) (*gen.PlannedPayment, error) {
	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return nil, err
	}
	if pp.Recurrence == "none" {
		return nil, &ValidationError{Field: "recurrence", Message: "cannot skip a one-time planned payment"}
	}

	currentDue := ""
	if pp.NextDueDate.Valid {
		currentDue = pp.NextDueDate.String
	} else {
		currentDue = pp.StartDate
	}
	dueTime, err := parseDate(currentDue)
	if err != nil {
		return nil, fmt.Errorf("parse current due date: %w", err)
	}
	dueDate, _ := time.Parse("2006-01-02", dueTime)
	newDue, err := calcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
	if err != nil {
		return nil, err
	}
	nextDueDate := newDue.Format("2006-01-02")

	updated, err := s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
		ID:          pp.ID,
		NextDueDate: sql.NullString{String: nextDueDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update next due date: %w", err)
	}

	return updated, nil
}

func (s *Service) PausePlannedPayment(id int64) error {
	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused != 0 {
		return &ValidationError{Field: "state", Message: "planned payment is already paused"}
	}
	return s.q.PausePlannedPayment(s.ctx(), id)
}

func (s *Service) ResumePlannedPayment(id int64) error {
	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused == 0 {
		return &ValidationError{Field: "state", Message: "planned payment is not paused"}
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	currentDue := ""
	if pp.NextDueDate.Valid {
		currentDue = pp.NextDueDate.String
	} else if pp.StartDate != "" {
		currentDue = pp.StartDate
	}

	if currentDue != "" && currentDue < today && pp.Recurrence != "none" {
		dueDate, _ := time.Parse("2006-01-02", currentDue)
		newDue, err := calcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
		if err != nil {
			return err
		}
		nextStr := newDue.Format("2006-01-02")
		for nextStr < today {
			newDue, err = calcNextDue(newDue, pp.Recurrence, pp.RecurrenceRule)
			if err != nil {
				return err
			}
			nextStr = newDue.Format("2006-01-02")
		}
		_, err = s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
			ID:          id,
			NextDueDate: sql.NullString{String: nextStr, Valid: true},
		})
		return err
	}

	return s.q.ResumePlannedPayment(s.ctx(), id)
}

func (s *Service) EditPlannedPayment(id int64, params EditPlannedPaymentParams) (*gen.PlannedPayment, error) {
	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return nil, err
	}

	var updateParams gen.UpdatePlannedPaymentParams
	updateParams.ID = id

	if params.Name != nil {
		updateParams.Name = sql.NullString{String: *params.Name, Valid: true}
	}
	if params.Amount != nil {
		if *params.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		updateParams.Amount = sql.NullInt64{Int64: *params.Amount, Valid: true}
	}
	if params.Currency != nil {
		updateParams.Currency = sql.NullString{String: *params.Currency, Valid: true}
	}
	if params.Type != nil {
		updateParams.Type = sql.NullString{String: *params.Type, Valid: true}
	}
	if params.Account != nil {
		account, err := s.ResolveAccount(*params.Account)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		updateParams.AccountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}
	if params.Category != nil {
		category, err := s.ResolveCategory(*params.Category)
		if err != nil {
			return nil, fmt.Errorf("category: %w", err)
		}
		updateParams.CategoryID = sql.NullInt64{Int64: category.ID, Valid: true}
	}
	if params.Recurrence != nil {
		validRecurrences := map[string]bool{
			"none": true, "daily": true, "weekly": true,
			"monthly": true, "yearly": true, "custom": true,
		}
		if !validRecurrences[*params.Recurrence] {
			return nil, &ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
		}
		if *params.Recurrence == "custom" {
			if params.RecurrenceRule == nil || *params.RecurrenceRule == "" {
				return nil, &ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
			}
			if err := validateRRULE(*params.RecurrenceRule); err != nil {
				return nil, &ValidationError{Field: "rrule", Message: err.Error()}
			}
		}
		updateParams.Recurrence = sql.NullString{String: *params.Recurrence, Valid: true}
	}
	if params.RecurrenceRule != nil {
		updateParams.RecurrenceRule = sql.NullString{String: *params.RecurrenceRule, Valid: true}
	}
	if params.StartDate != nil {
		parsed, err := parseDate(*params.StartDate)
		if err != nil {
			return nil, &ValidationError{Field: "start_date", Message: err.Error()}
		}
		updateParams.StartDate = sql.NullString{String: parsed, Valid: true}
	}

	dueDayChanged := params.DueDay != nil
	if dueDayChanged {
		effectiveRecurrence := pp.Recurrence
		if params.Recurrence != nil {
			effectiveRecurrence = *params.Recurrence
		}
		effectiveStart := pp.StartDate
		if updateParams.StartDate.Valid {
			effectiveStart = updateParams.StartDate.String
		}
		newDue := computeInitialDueDate(effectiveStart, effectiveRecurrence, *params.DueDay)
		updateParams.NextDueDate = sql.NullString{String: newDue, Valid: true}
	}

	updated, err := s.q.UpdatePlannedPayment(s.ctx(), updateParams)
	if err != nil {
		return nil, fmt.Errorf("update planned payment: %w", err)
	}

	return updated, nil
}

func (s *Service) DeletePlannedPayment(id int64) error {
	if _, err := s.GetPlannedPaymentByID(id); err != nil {
		return err
	}
	return s.q.DeletePlannedPayment(s.ctx(), id)
}

func computeInitialDueDate(startDate string, recurrence string, dueDay int) string {
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

func calcNextDue(currentDue time.Time, recurrence string, recurrenceRule sql.NullString) (time.Time, error) {
	switch recurrence {
	case "daily":
		return currentDue.AddDate(0, 0, 1), nil
	case "weekly":
		return currentDue.AddDate(0, 0, 7), nil
	case "monthly":
		day := currentDue.Day()
		nextMonth := currentDue.AddDate(0, 1, 0)
		lastDay := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, time.UTC), nil
	case "yearly":
		return currentDue.AddDate(1, 0, 0), nil
	case "custom":
		if !recurrenceRule.Valid {
			return time.Time{}, &ValidationError{Field: "recurrence_rule", Message: "custom recurrence rule is required"}
		}
		return calcNextDueFromRRULE(currentDue, recurrenceRule.String)
	case "none":
		return currentDue, nil
	default:
		return time.Time{}, &ValidationError{Field: "recurrence", Message: fmt.Sprintf("unknown recurrence: %s", recurrence)}
	}
}

func calcNextDueFromRRULE(currentDue time.Time, rrule string) (time.Time, error) {
	rule := strings.ToUpper(rrule)
	if !strings.HasPrefix(rule, "FREQ=") {
		return time.Time{}, &ValidationError{Field: "rrule", Message: "RRULE must start with FREQ="}
	}

	parts := strings.Split(rule, ";")
	freq := strings.TrimPrefix(parts[0], "FREQ=")
	var byMonthDay int

	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "BYMONTHDAY=") {
			fmt.Sscanf(part, "BYMONTHDAY=%d", &byMonthDay)
		}
	}

	switch freq {
	case "DAILY":
		return currentDue.AddDate(0, 0, 1), nil
	case "WEEKLY":
		return currentDue.AddDate(0, 0, 7), nil
	case "MONTHLY":
		day := currentDue.Day()
		if byMonthDay > 0 {
			day = byMonthDay
		}
		nextMonth := currentDue.AddDate(0, 1, 0)
		lastDay := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, time.UTC), nil
	case "YEARLY":
		return currentDue.AddDate(1, 0, 0), nil
	default:
		return time.Time{}, &ValidationError{Field: "rrule", Message: fmt.Sprintf("unsupported RRULE frequency: %s", freq)}
	}
}

func validateRRULE(rrule string) error {
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
