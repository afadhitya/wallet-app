package plannedpayment

import "github.com/afadhitya/wallet-app/internal/gen"

type PlannedPaymentManager struct {
	q gen.Querier
}

func NewPlannedPaymentManager(q gen.Querier) *PlannedPaymentManager {
	return &PlannedPaymentManager{q: q}
}
