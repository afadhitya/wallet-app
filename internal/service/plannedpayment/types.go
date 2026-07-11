package plannedpayment

import "github.com/afadhitya/wallet-app/internal/gen"

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
