package budget

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func CalculatePeriod(periodType, from, to string) (string, string, error) {
	now := time.Now()

	if from != "" && to != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return "", "", &shared.ValidationError{Field: "from", Message: "invalid date format (use YYYY-MM-DD)"}
		}
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return "", "", &shared.ValidationError{Field: "to", Message: "invalid date format (use YYYY-MM-DD)"}
		}
		return from, to, nil
	}

	switch periodType {
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	case "weekly":
		weekday := now.Weekday()
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 6)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	default:
		return "", "", &shared.ValidationError{Field: "period", Message: "one_time budget requires --from and --to dates"}
	}
}

func (m *BudgetManager) SetBudget(params SetBudgetParams) (*BudgetResult, error) {
	if params.Amount <= 0 {
		return nil, shared.ErrInvalidAmount
	}

	if len(params.Categories) == 0 && len(params.Tags) == 0 {
		return nil, &shared.ValidationError{Field: "targets", Message: "budget must have at least one category or tag target"}
	}

	if !validPeriods[params.Period] {
		return nil, &shared.ValidationError{Field: "period", Message: "supported periods: monthly, weekly, yearly, one_time"}
	}

	if params.NotifyPct == 0 {
		params.NotifyPct = 80
	}
	if params.NotifyPct < 1 || params.NotifyPct > 100 {
		return nil, &shared.ValidationError{Field: "notify", Message: "notification threshold must be between 1 and 100"}
	}

	periodStart, periodEnd, err := CalculatePeriod(params.Period, params.From, params.To)
	if err != nil {
		return nil, err
	}

	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}

	resolvedCategories := make([]*gen.Category, 0, len(params.Categories))
	for _, catName := range params.Categories {
		cat, err := shared.ResolveCategory(m.q, catName)
		if err != nil {
			return nil, fmt.Errorf("category '%s': %w", catName, err)
		}
		resolvedCategories = append(resolvedCategories, cat)
	}

	resolvedTags := make([]*gen.Tag, 0, len(params.Tags))
	for _, tagName := range params.Tags {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("tag '%s': %w", tagName, err)
		}
		resolvedTags = append(resolvedTags, tag)
	}

	ctx := context.Background()
	existing, err := m.q.GetBudgetByNameAndPeriod(ctx, gen.GetBudgetByNameAndPeriodParams{
		Name:        name,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing budget: %w", err)
	}
	if err == nil && existing != nil {
		return m.updateExistingBudget(ctx, existing.ID, params, periodTypeForDB(params.Period), periodStart, periodEnd, resolvedCategories, resolvedTags)
	}

	notifyPct := sql.NullInt64{Int64: params.NotifyPct, Valid: true}

	budget, err := m.q.CreateBudget(ctx, gen.CreateBudgetParams{
		Name:        name,
		Amount:      params.Amount,
		Currency:    "IDR",
		Type:        periodTypeForDB(params.Period),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		NotifyAtPct: notifyPct,
	})
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	for _, cat := range resolvedCategories {
		if err := m.q.AddBudgetCategory(ctx, gen.AddBudgetCategoryParams{
			BudgetID:   budget.ID,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, tag := range resolvedTags {
		if err := m.q.AddBudgetTag(ctx, gen.AddBudgetTagParams{
			BudgetID: budget.ID,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	return &BudgetResult{
		Budget:     budget,
		Categories: resolvedCategories,
		Tags:       resolvedTags,
	}, nil
}

func (m *BudgetManager) UpdateExistingBudget(id int64, params SetBudgetParams, periodType, periodStart, periodEnd string, categories []*gen.Category, tags []*gen.Tag) (*BudgetResult, error) {
	ctx := context.Background()
	return m.updateExistingBudget(ctx, id, params, periodType, periodStart, periodEnd, categories, tags)
}

func (m *BudgetManager) updateExistingBudget(ctx context.Context, id int64, params SetBudgetParams, periodType, periodStart, periodEnd string, categories []*gen.Category, tags []*gen.Tag) (*BudgetResult, error) {
	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}
	amountVal := sql.NullInt64{Int64: params.Amount, Valid: true}
	notifyVal := sql.NullInt64{Int64: params.NotifyPct, Valid: true}

	budget, err := m.q.UpdateBudget(ctx, gen.UpdateBudgetParams{
		ID:          id,
		Name:        name,
		Amount:      amountVal,
		NotifyAtPct: notifyVal,
		PeriodStart: sql.NullString{String: periodStart, Valid: true},
		PeriodEnd:   sql.NullString{String: periodEnd, Valid: true},
		Type:        sql.NullString{String: periodType, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}

	if err := m.q.RemoveAllBudgetCategories(ctx, id); err != nil {
		return nil, fmt.Errorf("remove budget categories: %w", err)
	}
	if err := m.q.RemoveAllBudgetTags(ctx, id); err != nil {
		return nil, fmt.Errorf("remove budget tags: %w", err)
	}

	for _, cat := range categories {
		if err := m.q.AddBudgetCategory(ctx, gen.AddBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, tag := range tags {
		if err := m.q.AddBudgetTag(ctx, gen.AddBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	return &BudgetResult{
		Budget:     budget,
		Categories: categories,
		Tags:       tags,
	}, nil
}

func periodTypeForDB(periodType string) string {
	if periodType == "one_time" {
		return "one_time"
	}
	return periodType
}
