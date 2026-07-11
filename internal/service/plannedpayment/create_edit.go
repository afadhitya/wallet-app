package plannedpayment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *PlannedPaymentManager) CreatePlannedPayment(params CreatePlannedPaymentParams) (*gen.PlannedPayment, error) {
	if params.Amount <= 0 {
		return nil, shared.ErrInvalidAmount
	}
	if params.Name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "name is required"}
	}
	if params.Currency == "" {
		params.Currency = "IDR"
	}
	if params.Type == "" {
		params.Type = "expense"
	}
	if params.Category == "" {
		return nil, &shared.ValidationError{Field: "category", Message: "category is required"}
	}

	validRecurrences := map[string]bool{
		"none": true, "daily": true, "weekly": true,
		"monthly": true, "yearly": true, "custom": true,
	}
	if !validRecurrences[params.Recurrence] {
		return nil, &shared.ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
	}
	if params.Recurrence == "none" && params.StartDate == "" {
		return nil, &shared.ValidationError{Field: "schedule", Message: "one-time planned payments require --from"}
	}

	if params.Recurrence == "custom" {
		if params.RecurrenceRule == "" {
			return nil, &shared.ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
		}
		if err := ValidateRRULE(params.RecurrenceRule); err != nil {
			return nil, &shared.ValidationError{Field: "rrule", Message: err.Error()}
		}
	}

	account, err := shared.ResolveAccount(m.q, params.Account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	category, err := shared.ResolveCategory(m.q, params.Category)
	if err != nil {
		return nil, fmt.Errorf("category: %w", err)
	}
	categoryID := sql.NullInt64{Int64: category.ID, Valid: true}

	startDate, err := shared.ParseDate(params.StartDate)
	if err != nil {
		return nil, &shared.ValidationError{Field: "start_date", Message: err.Error()}
	}

	var recurrenceRule sql.NullString
	if params.RecurrenceRule != "" {
		recurrenceRule = sql.NullString{String: params.RecurrenceRule, Valid: true}
	}

	nextDueDate := ComputeInitialDueDate(startDate, params.Recurrence, params.DueDay)
	var nextDueDateNull sql.NullString
	if nextDueDate != "" {
		nextDueDateNull = sql.NullString{String: nextDueDate, Valid: true}
	}

	ctx := context.Background()
	pp, err := m.q.CreatePlannedPayment(ctx, gen.CreatePlannedPaymentParams{
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

func (m *PlannedPaymentManager) GetPlannedPaymentByID(id int64) (*gen.PlannedPayment, error) {
	ctx := context.Background()
	pp, err := m.q.GetPlannedPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "planned payment", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}
	return pp, nil
}

func (m *PlannedPaymentManager) GetAccountName(id int64) (string, error) {
	ctx := context.Background()
	acc, err := m.q.GetAccountByID(ctx, id)
	if err != nil {
		return "", err
	}
	return acc.Name, nil
}

func (m *PlannedPaymentManager) GetCategoryName(id int64) string {
	ctx := context.Background()
	cat, err := m.q.GetCategoryByID(ctx, id)
	if err != nil {
		return ""
	}
	return cat.Name
}

func (m *PlannedPaymentManager) ListPlannedPayments(includePaused, includeArchived bool) ([]*gen.PlannedPayment, error) {
	ctx := context.Background()
	if includeArchived {
		if includePaused {
			return m.q.ListAllPlannedPayments(ctx)
		}
		active, err := m.q.ListActivePlannedPayments(ctx)
		if err != nil {
			return nil, err
		}
		archived, err := m.q.ListArchivedPlannedPayments(ctx)
		if err != nil {
			return nil, err
		}
		result := make([]*gen.PlannedPayment, 0, len(active)+len(archived))
		result = append(result, active...)
		result = append(result, archived...)
		return result, nil
	}
	if includePaused {
		return m.q.ListPausedPlannedPayments(ctx)
	}
	return m.q.ListActivePlannedPayments(ctx)
}

func (m *PlannedPaymentManager) ListDuePlannedPayments(params ListDueParams) ([]*gen.PlannedPayment, int64, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	var dateFrom, dateTo string

	switch params.Filter {
	case DueOverdue:
		ctx := context.Background()
		overdue, err := m.q.ListOverduePlannedPayments(ctx, sql.NullString{String: today, Valid: true})
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

	ctx := context.Background()
	due, err := m.q.ListDuePlannedPayments(ctx, gen.ListDuePlannedPaymentsParams{
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

func (m *PlannedPaymentManager) EditPlannedPayment(id int64, params EditPlannedPaymentParams) (*gen.PlannedPayment, error) {
	pp, err := m.GetPlannedPaymentByID(id)
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
			return nil, shared.ErrInvalidAmount
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
		account, err := shared.ResolveAccount(m.q, *params.Account)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		updateParams.AccountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}
	if params.Category != nil {
		category, err := shared.ResolveCategory(m.q, *params.Category)
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
			return nil, &shared.ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
		}
		if *params.Recurrence == "custom" {
			if params.RecurrenceRule == nil || *params.RecurrenceRule == "" {
				return nil, &shared.ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
			}
			if err := ValidateRRULE(*params.RecurrenceRule); err != nil {
				return nil, &shared.ValidationError{Field: "rrule", Message: err.Error()}
			}
		}
		updateParams.Recurrence = sql.NullString{String: *params.Recurrence, Valid: true}
	}
	if params.RecurrenceRule != nil {
		updateParams.RecurrenceRule = sql.NullString{String: *params.RecurrenceRule, Valid: true}
	}
	if params.StartDate != nil {
		parsed, err := shared.ParseDate(*params.StartDate)
		if err != nil {
			return nil, &shared.ValidationError{Field: "start_date", Message: err.Error()}
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
		newDue := ComputeInitialDueDate(effectiveStart, effectiveRecurrence, *params.DueDay)
		updateParams.NextDueDate = sql.NullString{String: newDue, Valid: true}
	}

	ctx := context.Background()
	updated, err := m.q.UpdatePlannedPayment(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("update planned payment: %w", err)
	}

	return updated, nil
}

func (m *PlannedPaymentManager) DeletePlannedPayment(id int64) error {
	if _, err := m.GetPlannedPaymentByID(id); err != nil {
		return err
	}
	ctx := context.Background()
	return m.q.DeletePlannedPayment(ctx, id)
}
