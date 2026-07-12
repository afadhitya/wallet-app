package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
	s.logger.Info("CreatePlannedPayment called", "name", params.Name, "amount", params.Amount, "currency", params.Currency, "type", params.Type)

	if params.Amount <= 0 {
		s.logger.Warn("CreatePlannedPayment invalid amount", "amount", params.Amount)
		return nil, ErrInvalidAmount
	}
	if params.Name == "" {
		s.logger.Warn("CreatePlannedPayment validation error", "field", "name", "message", "name is required")
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if params.Currency == "" {
		params.Currency = "IDR"
	}
	if params.Type == "" {
		params.Type = "expense"
	}
	if params.Category == "" {
		s.logger.Warn("CreatePlannedPayment validation error", "field", "category", "message", "category is required")
		return nil, &ValidationError{Field: "category", Message: "category is required"}
	}

	validRecurrences := map[string]bool{
		"none": true, "daily": true, "weekly": true,
		"monthly": true, "yearly": true, "custom": true,
	}
	if !validRecurrences[params.Recurrence] {
		s.logger.Warn("CreatePlannedPayment validation error", "field", "recurrence", "message", "recurrence must be one of: none, daily, weekly, monthly, yearly, custom")
		return nil, &ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
	}
	if params.Recurrence == "none" && params.StartDate == "" {
		s.logger.Warn("CreatePlannedPayment validation error", "field", "schedule", "message", "one-time planned payments require --from")
		return nil, &ValidationError{Field: "schedule", Message: "one-time planned payments require --from"}
	}

	if params.Recurrence == "custom" {
		if params.RecurrenceRule == "" {
			s.logger.Warn("CreatePlannedPayment validation error", "field", "rrule", "message", "recurrence rule is required when recurrence is custom")
			return nil, &ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
		}
		if err := validateRRULE(params.RecurrenceRule); err != nil {
			s.logger.Warn("CreatePlannedPayment validation error", "field", "rrule", "message", err.Error())
			return nil, &ValidationError{Field: "rrule", Message: err.Error()}
		}
	}

	account, err := s.ResolveAccount(params.Account)
	if err != nil {
		var notFound *NotFoundError
		var validation *ValidationError
		if errors.As(err, &notFound) || errors.As(err, &validation) {
			s.logger.Warn("CreatePlannedPayment account resolution failed", slog.String("error", err.Error()))
		} else {
			s.logger.Error("CreatePlannedPayment failed", slog.String("error", err.Error()))
		}
		return nil, fmt.Errorf("account: %w", err)
	}

	category, err := s.ResolveCategory(params.Category)
	if err != nil {
		var notFound *NotFoundError
		var validation *ValidationError
		if errors.As(err, &notFound) || errors.As(err, &validation) {
			s.logger.Warn("CreatePlannedPayment category resolution failed", slog.String("error", err.Error()))
		} else {
			s.logger.Error("CreatePlannedPayment failed", slog.String("error", err.Error()))
		}
		return nil, fmt.Errorf("category: %w", err)
	}
	categoryID := sql.NullInt64{Int64: category.ID, Valid: true}

	startDate, err := parseDate(params.StartDate)
	if err != nil {
		s.logger.Warn("CreatePlannedPayment validation error", "field", "start_date", "message", err.Error())
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
		s.logger.Error("CreatePlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create planned payment: %w", err)
	}

	s.logger.Info("CreatePlannedPayment completed", "id", pp.ID, "name", pp.Name)
	return pp, nil
}

func (s *Service) GetPlannedPaymentByID(id int64) (*gen.PlannedPayment, error) {
	s.logger.Info("GetPlannedPaymentByID called", "id", id)

	pp, err := s.q.GetPlannedPaymentByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetPlannedPaymentByID not found", "id", id)
			return nil, &NotFoundError{Entity: "planned payment", Name: fmt.Sprintf("%d", id)}
		}
		s.logger.Error("GetPlannedPaymentByID failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetPlannedPaymentByID completed", "id", id)
	return pp, nil
}

func (s *Service) GetAccountName(id int64) (string, error) {
	s.logger.Info("GetAccountName called", "id", id)

	acc, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		s.logger.Error("GetAccountName failed", slog.String("error", err.Error()))
		return "", err
	}
	s.logger.Info("GetAccountName completed", "id", id, "name", acc.Name)
	return acc.Name, nil
}

func (s *Service) GetCategoryName(id int64) string {
	s.logger.Info("GetCategoryName called", "id", id)

	cat, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		s.logger.Error("GetCategoryName failed", slog.String("error", err.Error()))
		return ""
	}
	s.logger.Info("GetCategoryName completed", "id", id, "name", cat.Name)
	return cat.Name
}

func (s *Service) ListPlannedPayments(includePaused, includeArchived bool) ([]*gen.PlannedPayment, error) {
	s.logger.Info("ListPlannedPayments called", "includePaused", includePaused, "includeArchived", includeArchived)

	if includeArchived {
		if includePaused {
			result, err := s.q.ListAllPlannedPayments(s.ctx())
			if err != nil {
				s.logger.Error("ListPlannedPayments failed", slog.String("error", err.Error()))
				return nil, err
			}
			s.logger.Info("ListPlannedPayments completed", "count", len(result))
			return result, nil
		}
		active, err := s.q.ListActivePlannedPayments(s.ctx())
		if err != nil {
			s.logger.Error("ListPlannedPayments failed", slog.String("error", err.Error()))
			return nil, err
		}
		archived, err := s.q.ListArchivedPlannedPayments(s.ctx())
		if err != nil {
			s.logger.Error("ListPlannedPayments failed", slog.String("error", err.Error()))
			return nil, err
		}
		result := make([]*gen.PlannedPayment, 0, len(active)+len(archived))
		result = append(result, active...)
		result = append(result, archived...)
		s.logger.Info("ListPlannedPayments completed", "count", len(result))
		return result, nil
	}
	if includePaused {
		result, err := s.q.ListPausedPlannedPayments(s.ctx())
		if err != nil {
			s.logger.Error("ListPlannedPayments failed", slog.String("error", err.Error()))
			return nil, err
		}
		s.logger.Info("ListPlannedPayments completed", "count", len(result))
		return result, nil
	}
	result, err := s.q.ListActivePlannedPayments(s.ctx())
	if err != nil {
		s.logger.Error("ListPlannedPayments failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListPlannedPayments completed", "count", len(result))
	return result, nil
}

func (s *Service) ListDuePlannedPayments(params ListDueParams) ([]*gen.PlannedPayment, int64, error) {
	s.logger.Info("ListDuePlannedPayments called", "filter", params.Filter, "nextDays", params.NextDays)

	now := time.Now()
	today := now.Format("2006-01-02")

	var dateFrom, dateTo string

	switch params.Filter {
	case DueOverdue:
		overdue, err := s.q.ListOverduePlannedPayments(s.ctx(), sql.NullString{String: today, Valid: true})
		if err != nil {
			s.logger.Error("ListDuePlannedPayments failed", slog.String("error", err.Error()))
			return nil, 0, err
		}
		var total int64
		for _, pp := range overdue {
			total += pp.Amount
		}
		s.logger.Info("ListDuePlannedPayments completed", "count", len(overdue), "total", total)
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
		s.logger.Error("ListDuePlannedPayments failed", slog.String("error", err.Error()))
		return nil, 0, err
	}

	var total int64
	for _, pp := range due {
		total += pp.Amount
	}
	s.logger.Info("ListDuePlannedPayments completed", "count", len(due), "total", total)
	return due, total, nil
}

func (s *Service) PayPlannedPayment(params PayPlannedPaymentParams) (*PayPlannedPaymentResult, error) {
	s.logger.Info("PayPlannedPayment called", "id", params.ID, "date", params.Date, "amount", params.Amount)

	pp, err := s.GetPlannedPaymentByID(params.ID)
	if err != nil {
		return nil, err
	}
	if pp.IsPaused != 0 {
		s.logger.Warn("PayPlannedPayment validation error", "field", "state", "message", "cannot pay a paused planned payment")
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
			s.logger.Warn("PayPlannedPayment validation error", "field", "date", "message", err.Error())
			return nil, &ValidationError{Field: "date", Message: err.Error()}
		}
	} else {
		date = time.Now().Format("2006-01-02")
	}

	account, err := s.q.GetAccountByID(s.ctx(), pp.AccountID)
	if err != nil {
		s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("account: %w", err)
	}

	var description sql.NullString
	if pp.Name != "" {
		description = sql.NullString{String: pp.Name, Valid: true}
	}

	txn, err := s.q.CreateTransaction(s.ctx(), gen.CreateTransactionParams{
		AccountID:    pp.AccountID,
		CategoryID:   pp.CategoryID,
		Type:         pp.Type,
		Amount:       amount,
		Currency:     pp.Currency,
		Description:  description,
		Notes:        sql.NullString{},
		TransferToID: sql.NullInt64{},
		Date:         date,
		BaseAmount:   sql.NullInt64{},
		BaseCurrency: sql.NullString{},
	})
	if err != nil {
		s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := s.recalculateBalance(account.ID); err != nil {
		s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	var nextDueDate string
	if pp.Recurrence == "none" {
		if err := s.q.ArchivePlannedPayment(s.ctx(), pp.ID); err != nil {
			s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
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
			s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("parse current due date: %w", err)
		}
		dueDate, _ := time.Parse("2006-01-02", dueTime)
		newDue, err := calcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
		if err != nil {
			s.logger.Warn("PayPlannedPayment calcNextDue failed", slog.String("error", err.Error()))
			return nil, err
		}
		nextDueDate = newDue.Format("2006-01-02")
		updated, err := s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
			ID:          pp.ID,
			NextDueDate: sql.NullString{String: nextDueDate, Valid: true},
		})
		if err != nil {
			s.logger.Error("PayPlannedPayment failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("update next due date: %w", err)
		}
		pp = updated
	}

	s.logger.Info("PayPlannedPayment completed", "transactionID", txn.ID, "plannedPaymentID", pp.ID)
	return &PayPlannedPaymentResult{
		Transaction:    txn,
		PlannedPayment: pp,
		NextDueDate:    nextDueDate,
	}, nil
}

func (s *Service) SkipPlannedPayment(id int64) (*gen.PlannedPayment, error) {
	s.logger.Info("SkipPlannedPayment called", "id", id)

	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return nil, err
	}
	if pp.Recurrence == "none" {
		s.logger.Warn("SkipPlannedPayment validation error", "field", "recurrence", "message", "cannot skip a one-time planned payment")
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
		s.logger.Error("SkipPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("parse current due date: %w", err)
	}
	dueDate, _ := time.Parse("2006-01-02", dueTime)
	newDue, err := calcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
	if err != nil {
		s.logger.Warn("SkipPlannedPayment calcNextDue failed", slog.String("error", err.Error()))
		return nil, err
	}
	nextDueDate := newDue.Format("2006-01-02")

	updated, err := s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
		ID:          pp.ID,
		NextDueDate: sql.NullString{String: nextDueDate, Valid: true},
	})
	if err != nil {
		s.logger.Error("SkipPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update next due date: %w", err)
	}

	s.logger.Info("SkipPlannedPayment completed", "id", id, "nextDueDate", nextDueDate)
	return updated, nil
}

func (s *Service) PausePlannedPayment(id int64) error {
	s.logger.Info("PausePlannedPayment called", "id", id)

	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused != 0 {
		s.logger.Warn("PausePlannedPayment validation error", "field", "state", "message", "planned payment is already paused")
		return &ValidationError{Field: "state", Message: "planned payment is already paused"}
	}
	err = s.q.PausePlannedPayment(s.ctx(), id)
	if err != nil {
		s.logger.Error("PausePlannedPayment failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("PausePlannedPayment completed", "id", id)
	return nil
}

func (s *Service) ResumePlannedPayment(id int64) error {
	s.logger.Info("ResumePlannedPayment called", "id", id)

	pp, err := s.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused == 0 {
		s.logger.Warn("ResumePlannedPayment validation error", "field", "state", "message", "planned payment is not paused")
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
			s.logger.Warn("ResumePlannedPayment calcNextDue failed", slog.String("error", err.Error()))
			return err
		}
		nextStr := newDue.Format("2006-01-02")
		for nextStr < today {
			newDue, err = calcNextDue(newDue, pp.Recurrence, pp.RecurrenceRule)
			if err != nil {
				s.logger.Warn("ResumePlannedPayment calcNextDue failed", slog.String("error", err.Error()))
				return err
			}
			nextStr = newDue.Format("2006-01-02")
		}
		if _, err := s.q.UpdatePlannedPaymentNextDueDate(s.ctx(), gen.UpdatePlannedPaymentNextDueDateParams{
			ID:          id,
			NextDueDate: sql.NullString{String: nextStr, Valid: true},
		}); err != nil {
			s.logger.Error("ResumePlannedPayment failed", slog.String("error", err.Error()))
			return err
		}
	}

	err = s.q.ResumePlannedPayment(s.ctx(), id)
	if err != nil {
		s.logger.Error("ResumePlannedPayment failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("ResumePlannedPayment completed", "id", id)
	return nil
}

func (s *Service) EditPlannedPayment(id int64, params EditPlannedPaymentParams) (*gen.PlannedPayment, error) {
	s.logger.Info("EditPlannedPayment called", "id", id)

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
			s.logger.Warn("EditPlannedPayment invalid amount", "amount", *params.Amount)
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
			var notFound *NotFoundError
			var validation *ValidationError
			if errors.As(err, &notFound) || errors.As(err, &validation) {
				s.logger.Warn("EditPlannedPayment account resolution failed", slog.String("error", err.Error()))
			} else {
				s.logger.Error("EditPlannedPayment failed", slog.String("error", err.Error()))
			}
			return nil, fmt.Errorf("account: %w", err)
		}
		updateParams.AccountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}
	if params.Category != nil {
		category, err := s.ResolveCategory(*params.Category)
		if err != nil {
			var notFound *NotFoundError
			var validation *ValidationError
			if errors.As(err, &notFound) || errors.As(err, &validation) {
				s.logger.Warn("EditPlannedPayment category resolution failed", slog.String("error", err.Error()))
			} else {
				s.logger.Error("EditPlannedPayment failed", slog.String("error", err.Error()))
			}
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
			s.logger.Warn("EditPlannedPayment validation error", "field", "recurrence", "message", "recurrence must be one of: none, daily, weekly, monthly, yearly, custom")
			return nil, &ValidationError{Field: "recurrence", Message: "recurrence must be one of: none, daily, weekly, monthly, yearly, custom"}
		}
		if *params.Recurrence == "custom" {
			if params.RecurrenceRule == nil || *params.RecurrenceRule == "" {
				s.logger.Warn("EditPlannedPayment validation error", "field", "rrule", "message", "recurrence rule is required when recurrence is custom")
				return nil, &ValidationError{Field: "rrule", Message: "recurrence rule is required when recurrence is custom"}
			}
			if err := validateRRULE(*params.RecurrenceRule); err != nil {
				s.logger.Warn("EditPlannedPayment validation error", "field", "rrule", "message", err.Error())
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
			s.logger.Warn("EditPlannedPayment validation error", "field", "start_date", "message", err.Error())
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
		s.logger.Error("EditPlannedPayment failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update planned payment: %w", err)
	}

	s.logger.Info("EditPlannedPayment completed", "id", id)
	return updated, nil
}

func (s *Service) DeletePlannedPayment(id int64) error {
	s.logger.Info("DeletePlannedPayment called", "id", id)

	if _, err := s.GetPlannedPaymentByID(id); err != nil {
		return err
	}
	err := s.q.DeletePlannedPayment(s.ctx(), id)
	if err != nil {
		s.logger.Error("DeletePlannedPayment failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("DeletePlannedPayment completed", "id", id)
	return nil
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
			return time.Time{}, &ValidationError{Field: "recurrence_rule", Message: "custom recurrence rule is required"}
		}
		return calcNextDueFromRRULE(currentDue, recurrenceRule.String)
	case "none":
		return currentDue, nil
	default:
		return time.Time{}, &ValidationError{Field: "recurrence", Message: fmt.Sprintf("unknown recurrence: %s", recurrence)}
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
		return time.Time{}, &ValidationError{Field: "rrule", Message: "RRULE must start with FREQ="}
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
