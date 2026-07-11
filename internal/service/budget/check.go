package budget

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *BudgetManager) ListBudgets(params ListBudgetsParams) ([]*BudgetListItem, error) {
	var budgets []*gen.Budget
	var err error

	ctx := context.Background()
	if params.All {
		budgets, err = m.q.ListAllBudgets(ctx)
	} else {
		budgets, err = m.q.ListActiveBudgets(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	items := make([]*BudgetListItem, 0, len(budgets))
	for _, b := range budgets {
		spent, err := m.calculateSpending(b.ID, b.PeriodStart, b.PeriodEnd)
		if err != nil {
			return nil, fmt.Errorf("calculate spending for budget %d: %w", b.ID, err)
		}
		remaining := b.Amount - spent

		categories, _ := m.q.ListBudgetCategories(ctx, b.ID)
		tags, _ := m.q.ListBudgetTags(ctx, b.ID)

		items = append(items, &BudgetListItem{
			Budget:     b,
			Spent:      spent,
			Remaining:  remaining,
			Categories: categories,
			Tags:       tags,
		})
	}
	return items, nil
}

func (m *BudgetManager) CheckBudgets(params CheckBudgetsParams) ([]*CheckBudgetResult, error) {
	if params.Identifier != "" {
		return m.checkSingleBudget(params.Identifier)
	}

	ctx := context.Background()
	budgets, err := m.q.ListActiveBudgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active budgets: %w", err)
	}

	results := make([]*CheckBudgetResult, 0, len(budgets))
	for _, b := range budgets {
		current, err := m.ensureCurrentPeriod(b)
		if err != nil {
			return nil, fmt.Errorf("ensure current period for budget '%s': %w", budgetName(b), err)
		}
		result, err := m.buildCheckResult(current)
		if err != nil {
			return nil, fmt.Errorf("build check result: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (m *BudgetManager) CheckSingleBudget(identifier string) ([]*CheckBudgetResult, error) {
	return m.checkSingleBudget(identifier)
}

func (m *BudgetManager) checkSingleBudget(identifier string) ([]*CheckBudgetResult, error) {
	budget, err := m.resolveBudget(identifier)
	if err != nil {
		return nil, err
	}

	current, err := m.ensureCurrentPeriod(budget)
	if err != nil {
		return nil, fmt.Errorf("ensure current period: %w", err)
	}

	result, err := m.buildCheckResult(current)
	if err != nil {
		return nil, fmt.Errorf("build check result: %w", err)
	}
	return []*CheckBudgetResult{result}, nil
}

func (m *BudgetManager) ResolveBudget(identifier string) (*gen.Budget, error) {
	ctx := context.Background()
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		budget, err := m.q.GetBudgetByID(ctx, id)
		if err == nil && budget.IsActive == 1 {
			return budget, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	budgets, err := m.q.ListActiveBudgets(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range budgets {
		if b.Name.Valid && b.Name.String == identifier {
			return b, nil
		}
	}
	return nil, &shared.NotFoundError{Entity: "budget", Name: identifier}
}

func (m *BudgetManager) resolveBudget(identifier string) (*gen.Budget, error) {
	return m.ResolveBudget(identifier)
}

func (m *BudgetManager) EnsureCurrentPeriod(budget *gen.Budget) (*gen.Budget, error) {
	return m.ensureCurrentPeriod(budget)
}

func (m *BudgetManager) ensureCurrentPeriod(budget *gen.Budget) (*gen.Budget, error) {
	if budget.Type == "one_time" {
		return budget, nil
	}

	now := time.Now()
	var currentStart, currentEnd string

	switch budget.Type {
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	case "weekly":
		weekday := now.Weekday()
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 6)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	default:
		return budget, nil
	}

	name := budget.Name
	ctx := context.Background()
	current, err := m.q.GetBudgetByNameAndPeriod(ctx, gen.GetBudgetByNameAndPeriodParams{
		Name:        name,
		PeriodStart: currentStart,
		PeriodEnd:   currentEnd,
	})
	if err == nil && current != nil {
		return current, nil
	}

	prior, err := m.q.GetMostRecentPriorBudget(ctx, gen.GetMostRecentPriorBudgetParams{
		Name:      name,
		PeriodEnd: currentStart,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return budget, nil
		}
		return nil, fmt.Errorf("get prior budget: %w", err)
	}

	newBudget, err := m.q.CreateBudget(ctx, gen.CreateBudgetParams{
		Name:        prior.Name,
		Amount:      prior.Amount,
		Currency:    prior.Currency,
		Type:        prior.Type,
		PeriodStart: currentStart,
		PeriodEnd:   currentEnd,
		NotifyAtPct: prior.NotifyAtPct,
	})
	if err != nil {
		return nil, fmt.Errorf("create recurring budget: %w", err)
	}

	categories, err := m.q.ListBudgetCategories(ctx, prior.ID)
	if err != nil {
		return nil, fmt.Errorf("list prior categories: %w", err)
	}
	for _, cat := range categories {
		if err := m.q.AddBudgetCategory(ctx, gen.AddBudgetCategoryParams{
			BudgetID:   newBudget.ID,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("copy budget category: %w", err)
		}
	}

	tags, err := m.q.ListBudgetTags(ctx, prior.ID)
	if err != nil {
		return nil, fmt.Errorf("list prior tags: %w", err)
	}
	for _, tag := range tags {
		if err := m.q.AddBudgetTag(ctx, gen.AddBudgetTagParams{
			BudgetID: newBudget.ID,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("copy budget tag: %w", err)
		}
	}

	return newBudget, nil
}

func (m *BudgetManager) BuildCheckResult(budget *gen.Budget) (*CheckBudgetResult, error) {
	return m.buildCheckResult(budget)
}

func (m *BudgetManager) buildCheckResult(budget *gen.Budget) (*CheckBudgetResult, error) {
	spent, err := m.calculateSpending(budget.ID, budget.PeriodStart, budget.PeriodEnd)
	if err != nil {
		return nil, err
	}
	remaining := budget.Amount - spent

	var percentUsed float64
	if budget.Amount > 0 {
		percentUsed = float64(spent) / float64(budget.Amount) * 100
	}

	notifyPct := int64(80)
	if budget.NotifyAtPct.Valid {
		notifyPct = budget.NotifyAtPct.Int64
	}

	status := BudgetStatusOK
	if percentUsed >= 100 {
		status = BudgetStatusOver
	} else if percentUsed >= float64(notifyPct) {
		status = BudgetStatusWarning
	}

	ctx := context.Background()
	categories, _ := m.q.ListBudgetCategories(ctx, budget.ID)
	tags, _ := m.q.ListBudgetTags(ctx, budget.ID)

	return &CheckBudgetResult{
		Budget:      budget,
		Spent:       spent,
		Remaining:   remaining,
		PercentUsed: percentUsed,
		Status:      status,
		Categories:  categories,
		Tags:        tags,
	}, nil
}
