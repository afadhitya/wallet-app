package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type CreateExpenseParams struct {
	Amount      int64
	Description string
	Category    string
	Account     string
	Tags        []string
	Date        string
	Notes       string
}

type CreateIncomeParams struct {
	Amount      int64
	Description string
	Category    string
	Account     string
	Tags        []string
	Date        string
	Notes       string
}

type TransactionResult struct {
	Transaction *gen.Transaction
	Tags        []*gen.Tag
}

func (s *Service) AddExpense(params CreateExpenseParams) (*TransactionResult, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	date, err := parseDate(params.Date)
	if err != nil {
		return nil, &ValidationError{Field: "date", Message: err.Error()}
	}

	category, err := s.ResolveCategory(params.Category)
	if err != nil {
		return nil, fmt.Errorf("category: %w", err)
	}

	account, err := s.ResolveAccount(params.Account)
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

	txn, err := s.queries.CreateTransaction(s.ctx(), gen.CreateTransactionParams{
		AccountID:   account.ID,
		CategoryID:  categoryID,
		Type:        "expense",
		Amount:      params.Amount,
		Currency:    "IDR",
		Description: description,
		Notes:       notes,
		Date:        date,
	})
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	tags := make([]*gen.Tag, 0)
	for _, tagName := range params.Tags {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("tag '%s': %w", tagName, err)
		}
		if err := s.AddTransactionTag(txn.ID, tag.ID); err != nil {
			return nil, fmt.Errorf("add tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := s.recalculateBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	return &TransactionResult{Transaction: txn, Tags: tags}, nil
}

func (s *Service) AddIncome(params CreateIncomeParams) (*TransactionResult, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	date, err := parseDate(params.Date)
	if err != nil {
		return nil, &ValidationError{Field: "date", Message: err.Error()}
	}

	category, err := s.ResolveCategory(params.Category)
	if err != nil {
		return nil, fmt.Errorf("category: %w", err)
	}

	account, err := s.ResolveAccount(params.Account)
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

	txn, err := s.queries.CreateTransaction(s.ctx(), gen.CreateTransactionParams{
		AccountID:   account.ID,
		CategoryID:  categoryID,
		Type:        "income",
		Amount:      params.Amount,
		Currency:    "IDR",
		Description: description,
		Notes:       notes,
		Date:        date,
	})
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	tags := make([]*gen.Tag, 0)
	for _, tagName := range params.Tags {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("tag '%s': %w", tagName, err)
		}
		if err := s.AddTransactionTag(txn.ID, tag.ID); err != nil {
			return nil, fmt.Errorf("add tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := s.recalculateBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	return &TransactionResult{Transaction: txn, Tags: tags}, nil
}

func (s *Service) recalculateBalance(accountID int64) error {
	balance, err := s.queries.GetAccountBalance(s.ctx(), accountID)
	if err != nil {
		return err
	}
	balanceInt, ok := balance.(int64)
	if !ok {
		return fmt.Errorf("unexpected balance type: %T", balance)
	}
	return s.UpdateAccountBalance(accountID, balanceInt)
}

func parseDate(input string) (string, error) {
	if input == "" {
		return time.Now().Format("2006-01-02"), nil
	}

	switch input {
	case "today":
		return time.Now().Format("2006-01-02"), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "tomorrow":
		return time.Now().AddDate(0, 0, 1).Format("2006-01-02"), nil
	}

	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("02/01/2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("02 Jan 2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse("2 Jan 2006", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", input)
}
