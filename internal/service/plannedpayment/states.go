package plannedpayment

import (
	"context"
	"database/sql"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *PlannedPaymentManager) PausePlannedPayment(id int64) error {
	pp, err := m.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused != 0 {
		return &shared.ValidationError{Field: "state", Message: "planned payment is already paused"}
	}
	ctx := context.Background()
	return m.q.PausePlannedPayment(ctx, id)
}

func (m *PlannedPaymentManager) ResumePlannedPayment(id int64) error {
	pp, err := m.GetPlannedPaymentByID(id)
	if err != nil {
		return err
	}
	if pp.IsPaused == 0 {
		return &shared.ValidationError{Field: "state", Message: "planned payment is not paused"}
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
		newDue, err := CalcNextDue(dueDate, pp.Recurrence, pp.RecurrenceRule)
		if err != nil {
			return err
		}
		nextStr := newDue.Format("2006-01-02")
		for nextStr < today {
			newDue, err = CalcNextDue(newDue, pp.Recurrence, pp.RecurrenceRule)
			if err != nil {
				return err
			}
			nextStr = newDue.Format("2006-01-02")
		}
		ctx := context.Background()
		if _, err := m.q.UpdatePlannedPaymentNextDueDate(ctx, gen.UpdatePlannedPaymentNextDueDateParams{
			ID:          id,
			NextDueDate: sql.NullString{String: nextStr, Valid: true},
		}); err != nil {
			return err
		}
	}

	ctx := context.Background()
	return m.q.ResumePlannedPayment(ctx, id)
}
