package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *Manager) AddTransfer(params CreateTransferParams) (*TransferResult, error) {
	if params.Amount <= 0 {
		return nil, shared.ErrInvalidAmount
	}

	if params.FromAccount == params.ToAccount {
		return nil, &shared.ValidationError{Field: "accounts", Message: "source and destination accounts must be different"}
	}

	date, err := shared.ParseDate(params.Date)
	if err != nil {
		return nil, &shared.ValidationError{Field: "date", Message: err.Error()}
	}

	fromAccount, err := shared.ResolveAccount(m.q, params.FromAccount)
	if err != nil {
		return nil, fmt.Errorf("source account: %w", err)
	}

	toAccount, err := shared.ResolveAccount(m.q, params.ToAccount)
	if err != nil {
		return nil, fmt.Errorf("destination account: %w", err)
	}

	ctx := context.Background()
	fromBalance, err := m.q.GetAccountBalance(ctx, fromAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("get source balance: %w", err)
	}
	fromBalanceInt := balanceToInt64(fromBalance)

	var warning string
	if fromBalanceInt < params.Amount {
		warning = fmt.Sprintf("Warning: insufficient balance in %s (balance: %d, transfer: %d)",
			fromAccount.Name, fromBalanceInt, params.Amount)
	}

	var description, notes sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}
	if params.Notes != "" {
		notes = sql.NullString{String: params.Notes, Valid: true}
	}

	transferToID := sql.NullInt64{Int64: toAccount.ID, Valid: true}

	txn, err := m.q.CreateTransaction(ctx, gen.CreateTransactionParams{
		AccountID:    fromAccount.ID,
		CategoryID:   sql.NullInt64{},
		Type:         "transfer",
		Amount:       params.Amount,
		Currency:     "IDR",
		Description:  description,
		Notes:        notes,
		TransferToID: transferToID,
		Date:         date,
	})
	if err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}

	if err := m.RecalcBalance(fromAccount.ID); err != nil {
		return nil, fmt.Errorf("recalculate source balance: %w", err)
	}
	if err := m.RecalcBalance(toAccount.ID); err != nil {
		return nil, fmt.Errorf("recalculate destination balance: %w", err)
	}

	return &TransferResult{Transaction: txn, Warning: warning}, nil
}

func BalanceToInt64(balance interface{}) int64 {
	if b, ok := balance.(int64); ok {
		return b
	}
	return 0
}

func balanceToInt64(balance interface{}) int64 {
	return BalanceToInt64(balance)
}
