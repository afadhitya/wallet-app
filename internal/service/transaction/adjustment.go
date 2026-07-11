package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *Manager) AdjustBalance(params AdjustBalanceParams) (*AdjustBalanceResult, error) {
	account, err := shared.ResolveAccount(m.q, params.Account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	ctx := context.Background()
	oldBalanceRaw, err := m.q.GetAccountBalance(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("get current balance: %w", err)
	}
	oldBalance := balanceToInt64(oldBalanceRaw)
	diff := params.Target - oldBalance

	if diff == 0 {
		return &AdjustBalanceResult{
			Account:    account,
			OldBalance: oldBalance,
			NewBalance: oldBalance,
			Difference: 0,
		}, nil
	}

	var description, notes sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}
	if params.Notes != "" {
		notes = sql.NullString{String: params.Notes, Valid: true}
	}

	txn, err := m.q.CreateTransaction(ctx, gen.CreateTransactionParams{
		AccountID:   account.ID,
		CategoryID:  sql.NullInt64{},
		Type:        "adjustment",
		Amount:      diff,
		Currency:    "IDR",
		Description: description,
		Notes:       notes,
		Date:        time.Now().Format("2006-01-02"),
	})
	if err != nil {
		return nil, fmt.Errorf("create adjustment: %w", err)
	}

	if err := m.RecalcBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	newBalanceRaw, err := m.q.GetAccountBalance(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("get new balance: %w", err)
	}
	newBalance := balanceToInt64(newBalanceRaw)

	return &AdjustBalanceResult{
		Account:     account,
		OldBalance:  oldBalance,
		NewBalance:  newBalance,
		Difference:  diff,
		Transaction: txn,
	}, nil
}
