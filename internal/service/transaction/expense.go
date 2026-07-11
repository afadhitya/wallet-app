package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *Manager) AddExpense(params CreateExpenseParams) (*TransactionResult, error) {
	if params.Amount <= 0 {
		return nil, shared.ErrInvalidAmount
	}

	date, err := shared.ParseDate(params.Date)
	if err != nil {
		return nil, &shared.ValidationError{Field: "date", Message: err.Error()}
	}

	category, err := shared.ResolveCategory(m.q, params.Category)
	if err != nil {
		return nil, fmt.Errorf("category: %w", err)
	}

	account, err := shared.ResolveAccount(m.q, params.Account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	var description sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}
	var notes sql.NullString
	if params.Notes != "" {
		notes = sql.NullString{String: params.Notes, Valid: true}
	}

	categoryID := sql.NullInt64{Int64: category.ID, Valid: true}

	baseAmount, baseCurrency, err := m.resolveBaseFields(account.Currency, params.Amount)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	txn, err := m.q.CreateTransaction(ctx, gen.CreateTransactionParams{
		AccountID:    account.ID,
		CategoryID:   categoryID,
		Type:         "expense",
		Amount:       params.Amount,
		Currency:     account.Currency,
		Description:  description,
		Notes:        notes,
		Date:         date,
		BaseAmount:   baseAmount,
		BaseCurrency: baseCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	tags := make([]*gen.Tag, 0)
	for _, tagName := range params.Tags {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("tag '%s': %w", tagName, err)
		}
		if err := m.q.AddTransactionTag(ctx, gen.AddTransactionTagParams{
			TransactionID: txn.ID,
			TagID:         tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := m.RecalcBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	return &TransactionResult{Transaction: txn, Tags: tags}, nil
}
