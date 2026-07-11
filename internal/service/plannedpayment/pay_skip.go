package plannedpayment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *PlannedPaymentManager) PayPlannedPayment(params PayPlannedPaymentParams) (*PayPlannedPaymentResult, error) {
	pp, err := m.GetPlannedPaymentByID(params.ID)
	if err != nil {
		return nil, err
	}
	if pp.IsPaused != 0 {
		return nil, &shared.ValidationError{Field: "state", Message: "cannot pay a paused planned payment"}
	}

	amount := pp.Amount
	if params.Amount > 0 {
		amount = params.Amount
	}

	date := ""
	if params.Date != "" {
		if parsed, err := shared.ParseDate(params.Date); err == nil {
			date = parsed
		} else {
			return nil, &shared.ValidationError{Field: "date", Message: err.Error()}
		}
	} else {
		date = time.Now().Format("2006-01-02")
	}

	ctx := context.Background()
	account, err := m.q.GetAccountByID(ctx, pp.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	var description sql.NullString
	if pp.Name != "" {
		description = sql.NullString{String: pp.Name, Valid: true}
	}

	txn, err := m.q.CreateTransaction(ctx, gen.CreateTransactionParams{
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
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := m.recalcBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	var nextDueDate string
	if pp.Recurrence == "none" {
		if err := m.q.ArchivePlannedPayment(ctx, pp.ID); err != nil {
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
		dueTime, err := shared.ParseDate(currentDue)
		if err != nil {
			return nil, fmt.Errorf("parse current due date: %w", err)
		}
		dueDate, _ := time.Parse("2006-01-02", dueTime)
		newDue, err := CalcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
		if err != nil {
			return nil, err
		}
		nextDueDate = newDue.Format("2006-01-02")
		updated, err := m.q.UpdatePlannedPaymentNextDueDate(ctx, gen.UpdatePlannedPaymentNextDueDateParams{
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

func (m *PlannedPaymentManager) SkipPlannedPayment(id int64) (*gen.PlannedPayment, error) {
	pp, err := m.GetPlannedPaymentByID(id)
	if err != nil {
		return nil, err
	}
	if pp.Recurrence == "none" {
		return nil, &shared.ValidationError{Field: "recurrence", Message: "cannot skip a one-time planned payment"}
	}

	currentDue := ""
	if pp.NextDueDate.Valid {
		currentDue = pp.NextDueDate.String
	} else {
		currentDue = pp.StartDate
	}
	dueTime, err := shared.ParseDate(currentDue)
	if err != nil {
		return nil, fmt.Errorf("parse current due date: %w", err)
	}
	dueDate, _ := time.Parse("2006-01-02", dueTime)
	newDue, err := CalcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
	if err != nil {
		return nil, err
	}
	nextDueDate := newDue.Format("2006-01-02")

	ctx := context.Background()
	updated, err := m.q.UpdatePlannedPaymentNextDueDate(ctx, gen.UpdatePlannedPaymentNextDueDateParams{
		ID:          pp.ID,
		NextDueDate: sql.NullString{String: nextDueDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update next due date: %w", err)
	}

	return updated, nil
}

func (m *PlannedPaymentManager) recalcBalance(accountID int64) error {
	ctx := context.Background()
	balance, err := m.q.GetAccountBalance(ctx, accountID)
	if err != nil {
		return err
	}
	balanceInt, ok := balance.(int64)
	if !ok {
		return fmt.Errorf("unexpected balance type: %T", balance)
	}
	return m.q.UpdateAccountBalance(ctx, gen.UpdateAccountBalanceParams{
		ID:      accountID,
		Balance: balanceInt,
	})
}
