package service

import (
	"database/sql"
	"errors"
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

type CreateTransferParams struct {
	Amount      int64
	FromAccount string
	ToAccount   string
	Date        string
	Description string
	Notes       string
}

type TransferResult struct {
	Transaction  *gen.Transaction
	Warning      string
}

func (s *Service) AddTransfer(params CreateTransferParams) (*TransferResult, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if params.FromAccount == params.ToAccount {
		return nil, &ValidationError{Field: "accounts", Message: "source and destination accounts must be different"}
	}

	date, err := parseDate(params.Date)
	if err != nil {
		return nil, &ValidationError{Field: "date", Message: err.Error()}
	}

	fromAccount, err := s.ResolveAccount(params.FromAccount)
	if err != nil {
		return nil, fmt.Errorf("source account: %w", err)
	}

	toAccount, err := s.ResolveAccount(params.ToAccount)
	if err != nil {
		return nil, fmt.Errorf("destination account: %w", err)
	}

	fromBalance, err := s.queries.GetAccountBalance(s.ctx(), fromAccount.ID)
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

	txn, err := s.queries.CreateTransaction(s.ctx(), gen.CreateTransactionParams{
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

	if err := s.recalculateBalance(fromAccount.ID); err != nil {
		return nil, fmt.Errorf("recalculate source balance: %w", err)
	}
	if err := s.recalculateBalance(toAccount.ID); err != nil {
		return nil, fmt.Errorf("recalculate destination balance: %w", err)
	}

	return &TransferResult{Transaction: txn, Warning: warning}, nil
}

func balanceToInt64(balance interface{}) int64 {
	if b, ok := balance.(int64); ok {
		return b
	}
	return 0
}

type ListTransactionsParams struct {
	AccountName  string
	CategoryName string
	TagName      string
	Type         string
	Month        string
	DateFrom     string
	DateTo       string
	Limit        int
	Offset       int
}

type ListTransactionsResult struct {
	Transactions []*gen.Transaction
	Total        int64
}

func (s *Service) ListTransactions(params ListTransactionsParams) (*ListTransactionsResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}

	var accountID, categoryID interface{}
	var tagName string
	var dateFrom, dateTo interface{}

	if params.AccountName != "" {
		account, err := s.ResolveAccount(params.AccountName)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		accountID = account.ID
	}

	if params.CategoryName != "" {
		category, err := s.ResolveCategory(params.CategoryName)
		if err != nil {
			return nil, fmt.Errorf("category: %w", err)
		}
		categoryID = category.ID
	}

	if params.TagName != "" {
		tagName = params.TagName
	}

	if params.DateFrom != "" {
		dateFrom = params.DateFrom
	}

	if params.DateTo != "" {
		dateTo = params.DateTo
	}

	if params.Month != "" {
		from, to, err := parseMonth(params.Month)
		if err != nil {
			return nil, fmt.Errorf("month: %w", err)
		}
		dateFrom = from
		dateTo = to
	}

	if dateFrom == nil && dateTo == nil && params.Month == "" && params.DateFrom == "" && params.DateTo == "" {
		now := time.Now()
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, -1)
		dateFrom = firstDay.Format("2006-01-02")
		dateTo = lastDay.Format("2006-01-02")
	}

	if tagName != "" {
		transactions, err := s.queries.ListTransactionsByTag(s.ctx(), gen.ListTransactionsByTagParams{
			TagName:    tagName,
			AccountID:  accountID,
			CategoryID: categoryID,
			Type:       stringToInterface(params.Type),
			DateFrom:   dateFrom,
			DateTo:     dateTo,
			Limit:      int64(params.Limit),
			Offset:     int64(params.Offset),
		})
		if err != nil {
			return nil, err
		}
		result := &ListTransactionsResult{Transactions: transactions}
		result.Total = s.sumTransactionAmounts(transactions)
		return result, nil
	}

	transactions, err := s.queries.ListTransactions(s.ctx(), gen.ListTransactionsParams{
		AccountID:  accountID,
		CategoryID: categoryID,
		Type:       stringToInterface(params.Type),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Limit:      int64(params.Limit),
		Offset:     int64(params.Offset),
	})
	if err != nil {
		return nil, err
	}

	result := &ListTransactionsResult{Transactions: transactions}
	result.Total = s.sumTransactionAmounts(transactions)
	return result, nil
}

func (s *Service) sumTransactionAmounts(transactions []*gen.Transaction) int64 {
	var total int64
	for _, t := range transactions {
		total += t.Amount
	}
	return total
}

func stringToInterface(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func parseMonth(input string) (string, string, error) {
	now := time.Now()

	months := map[string]time.Month{
		"january": 1, "february": 2, "march": 3, "april": 4,
		"may": 5, "june": 6, "july": 7, "august": 8,
		"september": 9, "october": 10, "november": 11, "december": 12,
		"jan": 1, "feb": 2, "mar": 3, "apr": 4,
		"jun": 6, "jul": 7, "aug": 8,
		"sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}

	month, ok := months[input]
	if !ok {
		if t, err := time.Parse("2006-01", input); err == nil {
			month = t.Month()
			now = time.Date(t.Year(), month, 1, 0, 0, 0, 0, time.UTC)
		} else if t, err := time.Parse("01/2006", input); err == nil {
			month = t.Month()
			now = time.Date(t.Year(), month, 1, 0, 0, 0, 0, time.UTC)
		} else {
			return "", "", fmt.Errorf("invalid month: %s", input)
		}
	}

	firstDay := time.Date(now.Year(), month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), nil
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

type EditTransactionParams struct {
	Amount         *int64
	CategoryName   string
	AccountName    string
	Date           string
	Description    string
	Notes          string
	AddTagNames    []string
	RemoveTagNames []string
}

func (s *Service) EditTransaction(id int64, params EditTransactionParams) (*TransactionResult, error) {
	oldTxn, err := s.queries.GetTransactionByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	var amountVal sql.NullInt64
	if params.Amount != nil {
		if *params.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		amountVal = sql.NullInt64{Int64: *params.Amount, Valid: true}
	}

	var categoryID sql.NullInt64
	if params.CategoryName != "" {
		category, err := s.ResolveCategory(params.CategoryName)
		if err != nil {
			return nil, fmt.Errorf("category: %w", err)
		}
		categoryID = sql.NullInt64{Int64: category.ID, Valid: true}
	}

	var accountID sql.NullInt64
	if params.AccountName != "" {
		account, err := s.ResolveAccount(params.AccountName)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		accountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}

	var dateVal sql.NullString
	if params.Date != "" {
		parsed, err := parseDate(params.Date)
		if err != nil {
			return nil, &ValidationError{Field: "date", Message: err.Error()}
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

	updated, err := s.queries.UpdateTransaction(s.ctx(), gen.UpdateTransactionParams{
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
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("add tag '%s': %w", tagName, err)
		}
		if err := s.AddTransactionTag(id, tag.ID); err != nil {
			return nil, fmt.Errorf("add tag: %w", err)
		}
	}

	for _, tagName := range params.RemoveTagNames {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("remove tag '%s': %w", tagName, err)
		}
		if err := s.RemoveTransactionTag(id, tag.ID); err != nil {
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
		if err := s.recalculateBalance(acctID); err != nil {
			return nil, fmt.Errorf("recalculate balance for account %d: %w", acctID, err)
		}
	}

	tags, err := s.ListTransactionTags(id)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	return &TransactionResult{Transaction: updated, Tags: tags}, nil
}

func (s *Service) RemoveTransaction(id int64) error {
	txn, err := s.queries.GetTransactionByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return err
	}

	if err := s.queries.ArchiveTransaction(s.ctx(), id); err != nil {
		return fmt.Errorf("archive transaction: %w", err)
	}

	affectedAccounts := make(map[int64]bool)
	affectedAccounts[txn.AccountID] = true
	if txn.Type == "transfer" && txn.TransferToID.Valid {
		affectedAccounts[txn.TransferToID.Int64] = true
	}

	for acctID := range affectedAccounts {
		if err := s.recalculateBalance(acctID); err != nil {
			return fmt.Errorf("recalculate balance for account %d: %w", acctID, err)
		}
	}

	return nil
}

type AdjustBalanceParams struct {
	Account     string
	Target      int64
	Description string
	Notes       string
}

type AdjustBalanceResult struct {
	Account      *gen.Account
	OldBalance   int64
	NewBalance   int64
	Difference   int64
	Transaction  *gen.Transaction
}

func (s *Service) AdjustBalance(params AdjustBalanceParams) (*AdjustBalanceResult, error) {
	account, err := s.ResolveAccount(params.Account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	oldBalanceRaw, err := s.queries.GetAccountBalance(s.ctx(), account.ID)
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

	txn, err := s.queries.CreateTransaction(s.ctx(), gen.CreateTransactionParams{
		AccountID:  account.ID,
		CategoryID: sql.NullInt64{},
		Type:       "adjustment",
		Amount:     diff,
		Currency:   "IDR",
		Description: description,
		Notes:      notes,
		Date:       time.Now().Format("2006-01-02"),
	})
	if err != nil {
		return nil, fmt.Errorf("create adjustment: %w", err)
	}

	if err := s.recalculateBalance(account.ID); err != nil {
		return nil, fmt.Errorf("recalculate balance: %w", err)
	}

	newBalanceRaw, err := s.queries.GetAccountBalance(s.ctx(), account.ID)
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

func (s *Service) GetTransactionByID(id int64) (*gen.Transaction, error) {
	txn, err := s.queries.GetTransactionByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "transaction", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}
	return txn, nil
}
