package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *Manager) EditTransaction(id int64, params EditTransactionParams) (*TransactionResult, error) {
	ctx := context.Background()
	oldTxn, err := m.q.GetTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	var amountVal sql.NullInt64
	if params.Amount != nil {
		if *params.Amount <= 0 {
			return nil, shared.ErrInvalidAmount
		}
		amountVal = sql.NullInt64{Int64: *params.Amount, Valid: true}
	}

	var categoryID sql.NullInt64
	if params.CategoryName != "" {
		category, err := shared.ResolveCategory(m.q, params.CategoryName)
		if err != nil {
			return nil, fmt.Errorf("category: %w", err)
		}
		categoryID = sql.NullInt64{Int64: category.ID, Valid: true}
	}

	var accountID sql.NullInt64
	if params.AccountName != "" {
		account, err := shared.ResolveAccount(m.q, params.AccountName)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		accountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}

	var dateVal sql.NullString
	if params.Date != "" {
		parsed, err := shared.ParseDate(params.Date)
		if err != nil {
			return nil, &shared.ValidationError{Field: "date", Message: err.Error()}
		}
		dateVal = sql.NullString{String: parsed, Valid: true}
	}

	var descVal, notesVal sql.NullString
	if params.Description != "" {
		descVal = sql.NullString{String: params.Description, Valid: true}
	}
	if params.Notes != "" {
		notesVal = sql.NullString{String: params.Notes, Valid: true}
	}

	updated, err := m.q.UpdateTransaction(ctx, gen.UpdateTransactionParams{
		ID:          id,
		Amount:      amountVal,
		CategoryID:  categoryID,
		AccountID:   accountID,
		Date:        dateVal,
		Description: descVal,
		Notes:       notesVal,
	})
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	for _, tagName := range params.AddTagNames {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("add tag '%s': %w", tagName, err)
		}
		if err := m.q.AddTransactionTag(ctx, gen.AddTransactionTagParams{
			TransactionID: id,
			TagID:         tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add tag: %w", err)
		}
	}

	for _, tagName := range params.RemoveTagNames {
		tag, err := shared.ResolveTag(m.q, tagName)
		if err != nil {
			return nil, fmt.Errorf("remove tag '%s': %w", tagName, err)
		}
		if err := m.q.RemoveTransactionTag(ctx, gen.RemoveTransactionTagParams{
			TransactionID: id,
			TagID:         tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("remove tag: %w", err)
		}
	}

	affectedAccounts := make(map[int64]bool)
	affectedAccounts[oldTxn.AccountID] = true
	if updated.AccountID != oldTxn.AccountID {
		affectedAccounts[updated.AccountID] = true
	}
	if oldTxn.Type == "transfer" {
		if oldTxn.TransferToID.Valid {
			affectedAccounts[oldTxn.TransferToID.Int64] = true
		}
	}
	if updated.Type == "transfer" {
		if updated.TransferToID.Valid {
			affectedAccounts[updated.TransferToID.Int64] = true
		}
	}

	for acctID := range affectedAccounts {
		if err := m.RecalcBalance(acctID); err != nil {
			return nil, fmt.Errorf("recalculate balance for account %d: %w", acctID, err)
		}
	}

	tags, err := m.q.ListTransactionTags(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	return &TransactionResult{Transaction: updated, Tags: tags}, nil
}

func (m *Manager) RemoveTransaction(id int64) error {
	ctx := context.Background()
	txn, err := m.q.GetTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &shared.NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return err
	}

	if err := m.q.ArchiveTransaction(ctx, id); err != nil {
		return fmt.Errorf("archive transaction: %w", err)
	}

	affectedAccounts := make(map[int64]bool)
	affectedAccounts[txn.AccountID] = true
	if txn.Type == "transfer" && txn.TransferToID.Valid {
		affectedAccounts[txn.TransferToID.Int64] = true
	}

	for acctID := range affectedAccounts {
		if err := m.RecalcBalance(acctID); err != nil {
			return fmt.Errorf("recalculate balance for account %d: %w", acctID, err)
		}
	}

	return nil
}

func (m *Manager) GetTransactionByID(id int64) (*gen.Transaction, error) {
	ctx := context.Background()
	txn, err := m.q.GetTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}
	return txn, nil
}

func (m *Manager) RecalcBalance(accountID int64) error {
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

func (m *Manager) ResolveBaseFields(accountCurrency string, amount int64) (sql.NullInt64, sql.NullString, error) {
	return m.resolveBaseFields(accountCurrency, amount)
}

func (m *Manager) resolveBaseFields(accountCurrency string, amount int64) (sql.NullInt64, sql.NullString, error) {
	baseCurrency, err := shared.GetBaseCurrency()
	if err != nil {
		return sql.NullInt64{}, sql.NullString{}, err
	}

	if accountCurrency == baseCurrency {
		return sql.NullInt64{}, sql.NullString{}, nil
	}

	converted, err := shared.Convert(amount, accountCurrency)
	if err != nil {
		return sql.NullInt64{}, sql.NullString{}, err
	}

	return sql.NullInt64{Int64: converted, Valid: true}, sql.NullString{String: baseCurrency, Valid: true}, nil
}
