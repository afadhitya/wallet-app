package budget

import "github.com/afadhitya/wallet-app/internal/gen"

type BudgetManager struct {
	q gen.Querier
}

func NewBudgetManager(q gen.Querier) *BudgetManager {
	return &BudgetManager{q: q}
}
