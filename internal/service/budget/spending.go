package budget

import (
	"context"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (m *BudgetManager) CalculateSpending(budgetID int64, periodStart, periodEnd string) (int64, error) {
	return m.calculateSpending(budgetID, periodStart, periodEnd)
}

func (m *BudgetManager) calculateSpending(budgetID int64, periodStart, periodEnd string) (int64, error) {
	ctx := context.Background()
	params := gen.SumCategoryExpensesParams{
		BudgetID:    budgetID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
	catTotal, err := m.q.SumCategoryExpenses(ctx, params)
	if err != nil {
		return 0, err
	}
	catAmount := toInt64(catTotal)

	tagParams := gen.SumTagExpensesParams{
		BudgetID:    budgetID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
	tagTotal, err := m.q.SumTagExpenses(ctx, tagParams)
	if err != nil {
		return 0, err
	}
	tagAmount := toInt64(tagTotal)

	return catAmount + tagAmount, nil
}

func ToInt64(v interface{}) int64 {
	if b, ok := v.(int64); ok {
		return b
	}
	return 0
}

func toInt64(v interface{}) int64 {
	return ToInt64(v)
}
