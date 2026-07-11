package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (m *Manager) ListTransactions(params ListTransactionsParams) (*ListTransactionsResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}

	var accountID, categoryID interface{}
	var tagName string
	var dateFrom, dateTo interface{}

	if params.AccountName != "" {
		account, err := shared.ResolveAccount(m.q, params.AccountName)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}
		accountID = account.ID
	}

	if params.CategoryName != "" {
		category, err := shared.ResolveCategory(m.q, params.CategoryName)
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
		from, to, err := shared.ParseMonth(params.Month)
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

	ctx := context.Background()
	if tagName != "" {
		transactions, err := m.q.ListTransactionsByTag(ctx, gen.ListTransactionsByTagParams{
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
		result := buildListResult(m, transactions)
		return result, nil
	}

	transactions, err := m.q.ListTransactions(ctx, gen.ListTransactionsParams{
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

	result := buildListResult(m, transactions)
	return result, nil
}

func buildListResult(m *Manager, transactions []*gen.Transaction) *ListTransactionsResult {
	result := &ListTransactionsResult{Transactions: transactions}
	result.Total = sumTransactionAmounts(transactions)
	result.BaseTotal = sumBaseAmounts(transactions)
	return result
}

func sumBaseAmounts(transactions []*gen.Transaction) int64 {
	hasAnyBase := false
	var baseTotal int64
	for _, t := range transactions {
		if t.BaseAmount.Valid {
			hasAnyBase = true
			baseTotal += t.BaseAmount.Int64
		} else {
			baseTotal += t.Amount
		}
	}
	if !hasAnyBase {
		return 0
	}
	return baseTotal
}

func SumTransactionAmounts(transactions []*gen.Transaction) int64 {
	return sumTransactionAmounts(transactions)
}

func sumTransactionAmounts(transactions []*gen.Transaction) int64 {
	var total int64
	for _, t := range transactions {
		total += t.Amount
	}
	return total
}

func StringToInterface(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func stringToInterface(s string) interface{} {
	return StringToInterface(s)
}
