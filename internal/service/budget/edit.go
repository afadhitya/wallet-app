package budget

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *BudgetManager) EditBudget(id int64, params EditBudgetParams) (*BudgetResult, error) {
	ctx := context.Background()
	_, err := m.q.GetBudgetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "budget", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}

	var amountVal sql.NullInt64
	if params.Amount != nil {
		if *params.Amount <= 0 {
			return nil, shared.ErrInvalidAmount
		}
		amountVal = sql.NullInt64{Int64: *params.Amount, Valid: true}
	}

	var notifyVal sql.NullInt64
	if params.NotifyPct != nil {
		if *params.NotifyPct < 1 || *params.NotifyPct > 100 {
			return nil, &shared.ValidationError{Field: "notify", Message: "notification threshold must be between 1 and 100"}
		}
		notifyVal = sql.NullInt64{Int64: *params.NotifyPct, Valid: true}
	}

	updated, err := m.q.UpdateBudget(ctx, gen.UpdateBudgetParams{
		ID:          id,
		Name:        name,
		Amount:      amountVal,
		NotifyAtPct: notifyVal,
	})
	if err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}

	for _, catName := range params.AddCategories {
		cat, err := shared.ResolveCategory(m.q, catName)
		if err != nil {
			return nil, fmt.Errorf("add category '%s': %w", catName, err)
		}
		if err := m.q.AddBudgetCategory(ctx, gen.AddBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, catName := range params.RemoveCategories {
		cat, err := shared.ResolveCategory(m.q, catName)
		if err != nil {
			return nil, fmt.Errorf("remove category '%s': %w", catName, err)
		}
		if err := m.q.RemoveBudgetCategory(ctx, gen.RemoveBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("remove budget category: %w", err)
		}
	}

	for _, tagName := range params.AddTags {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("add tag '%s': %w", tagName, err)
		}
		if err := m.q.AddBudgetTag(ctx, gen.AddBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	for _, tagName := range params.RemoveTags {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("remove tag '%s': %w", tagName, err)
		}
		if err := m.q.RemoveBudgetTag(ctx, gen.RemoveBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("remove budget tag: %w", err)
		}
	}

	categories, _ := m.q.ListBudgetCategories(ctx, updated.ID)
	tags, _ := m.q.ListBudgetTags(ctx, updated.ID)

	return &BudgetResult{
		Budget:     updated,
		Categories: categories,
		Tags:       tags,
	}, nil
}

func (m *BudgetManager) RemoveBudget(id int64) error {
	ctx := context.Background()
	_, err := m.q.GetBudgetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &shared.NotFoundError{Entity: "budget", Name: fmt.Sprintf("%d", id)}
		}
		return err
	}
	return m.q.MarkBudgetInactive(ctx, id)
}

func BudgetName(b *gen.Budget) string {
	if b.Name.Valid {
		return b.Name.String
	}
	return fmt.Sprintf("%d", b.ID)
}

func budgetName(b *gen.Budget) string {
	return BudgetName(b)
}
