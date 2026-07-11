package report

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

var (
	ErrInvalidMonth  = &shared.ValidationError{Field: "month", Message: "Invalid month format. Expected month name or YYYY-MM."}
	ErrInvalidExport = &shared.ValidationError{Field: "export", Message: "Unsupported export format. Only 'csv' is supported."}
	ErrInvalidBy     = &shared.ValidationError{Field: "by", Message: "Unsupported breakdown. Expected 'category', 'account', or 'tag'."}
	ErrExportFailed  = fmt.Errorf("export failed")
	ErrNoReportData  = fmt.Errorf("no transactions found for specified period")
)

func (m *ReportManager) GenerateReport(params ReportParams) (*ReportResult, error) {
	baseCurrency, err := shared.GetBaseCurrency()
	if err != nil {
		return nil, err
	}

	filters, err := m.resolveReportFilters(params)
	if err != nil {
		return nil, err
	}

	switch params.By {
	case "":
		return m.generateMonthlySummary(baseCurrency, filters)
	case "category":
		return m.generateCategoryBreakdownResult(baseCurrency, filters)
	case "account":
		return m.generateAccountBreakdownResult(baseCurrency, filters)
	case "tag":
		return m.generateTagBreakdownResult(baseCurrency, filters)
	default:
		return nil, ErrInvalidBy
	}
}

func (m *ReportManager) resolveReportFilters(params ReportParams) (ReportFilters, error) {
	var filters ReportFilters

	period, dateFrom, dateTo, err := m.resolvePeriod(params)
	if err != nil {
		return filters, err
	}
	filters.Period = period
	filters.DateFrom = dateFrom
	filters.DateTo = dateTo

	if params.AccountName != "" {
		account, err := shared.ResolveAccount(m.q, params.AccountName)
		if err != nil {
			return filters, err
		}
		filters.AccountID = &account.ID
	}

	return filters, nil
}

func (m *ReportManager) resolvePeriod(params ReportParams) (string, string, string, error) {
	now := time.Now()

	if params.DateFrom != "" || params.DateTo != "" {
		if params.DateFrom == "" {
			return "", "", "", &shared.ValidationError{Field: "from", Message: "date range requires both --from and --to"}
		}
		if params.DateTo == "" {
			return "", "", "", &shared.ValidationError{Field: "to", Message: "date range requires both --from and --to"}
		}
		if _, err := time.Parse("2006-01-02", params.DateFrom); err != nil {
			return "", "", "", &shared.ValidationError{Field: "from", Message: "date must be YYYY-MM-DD"}
		}
		if _, err := time.Parse("2006-01-02", params.DateTo); err != nil {
			return "", "", "", &shared.ValidationError{Field: "to", Message: "date must be YYYY-MM-DD"}
		}
		period := params.DateFrom + " to " + params.DateTo
		return period, params.DateFrom, params.DateTo, nil
	}

	if params.Month != "" {
		from, to, err := shared.ParseMonth(params.Month)
		if err != nil {
			return "", "", "", ErrInvalidMonth
		}

		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			return "", "", "", err
		}
		period := t.Format("January 2006")
		return period, from, to, nil
	}

	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	period := now.Format("January 2006")
	return period, firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), nil
}

func nullFloatToInt64(v sql.NullFloat64) int64 {
	if v.Valid {
		return int64(math.Round(v.Float64))
	}
	return 0
}

func genInterfaceToInt64(v interface{}) int64 {
	if b, ok := v.(int64); ok {
		return b
	}
	return 0
}

func (m *ReportManager) genReportParams(filters ReportFilters) gen.ReportByAccountParams {
	var accountID interface{}
	if filters.AccountID != nil {
		accountID = *filters.AccountID
	}
	return gen.ReportByAccountParams{
		DateFrom:  filters.DateFrom,
		DateTo:    filters.DateTo,
		AccountID: accountID,
	}
}

func (m *ReportManager) generateMonthlySummary(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := m.genReportParams(filters)
	ctx := context.Background()

	incomeTotal, err := m.q.ReportIncomeTotal(ctx, gen.ReportIncomeTotalParams(arg))
	if err != nil {
		return nil, err
	}

	expenseTotal, err := m.q.ReportExpenseTotal(ctx, gen.ReportExpenseTotalParams(arg))
	if err != nil {
		return nil, err
	}

	transfers, err := m.q.ReportTransfers(ctx, gen.ReportTransfersParams(arg))
	if err != nil {
		return nil, err
	}

	txnCount, err := m.q.ReportTransactionCount(ctx, gen.ReportTransactionCountParams(arg))
	if err != nil {
		return nil, err
	}

	incomeTotalVal := genInterfaceToInt64(incomeTotal)
	expenseTotalVal := genInterfaceToInt64(expenseTotal)
	transferTotal := genInterfaceToInt64(transfers)

	incomeRows, err := m.q.ReportIncomeByCategory(ctx, gen.ReportIncomeByCategoryParams(arg))
	if err != nil {
		return nil, err
	}

	expenseRows, err := m.q.ReportExpenseByCategory(ctx, gen.ReportExpenseByCategoryParams(arg))
	if err != nil {
		return nil, err
	}

	incomeCategories := make([]ReportCategoryRow, 0, len(incomeRows))
	for _, r := range incomeRows {
		amount := nullFloatToInt64(r.Total)
		incomeCategories = append(incomeCategories, ReportCategoryRow{
			CategoryID:       r.CategoryID,
			CategoryName:     r.CategoryName,
			Amount:           amount,
			TransactionCount: r.TransactionCount,
		})
	}

	expenseCategories := make([]ReportCategoryRow, 0, len(expenseRows))
	for _, r := range expenseRows {
		amount := nullFloatToInt64(r.Total)
		parentName := ""
		if r.ParentCategoryName.Valid {
			parentName = r.ParentCategoryName.String
		}
		expenseCategories = append(expenseCategories, ReportCategoryRow{
			CategoryID:         r.CategoryID,
			CategoryName:       r.CategoryName,
			ParentCategoryName: parentName,
			Amount:             amount,
			TransactionCount:   r.TransactionCount,
		})
	}

	return &ReportResult{
		BaseCurrency:      baseCurrency,
		Period:            filters.Period,
		IncomeTotal:       incomeTotalVal,
		ExpenseTotal:      expenseTotalVal,
		Net:               incomeTotalVal - expenseTotalVal,
		TransferTotal:     transferTotal,
		TransactionCount:  txnCount,
		IncomeCategories:  incomeCategories,
		ExpenseCategories: expenseCategories,
	}, nil
}

func (m *ReportManager) generateCategoryBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := m.genReportParams(filters)
	ctx := context.Background()

	expenseTotal, err := m.q.ReportExpenseTotal(ctx, gen.ReportExpenseTotalParams(arg))
	if err != nil {
		return nil, err
	}
	expenseTotalVal := genInterfaceToInt64(expenseTotal)

	rows, err := m.q.ReportExpenseByCategory(ctx, gen.ReportExpenseByCategoryParams(arg))
	if err != nil {
		return nil, err
	}

	txnCount, err := m.q.ReportTransactionCount(ctx, gen.ReportTransactionCountParams(arg))
	if err != nil {
		return nil, err
	}

	categories := make([]ReportCategoryRow, 0, len(rows))
	for _, r := range rows {
		amount := nullFloatToInt64(r.Total)
		pct := float64(0)
		if expenseTotalVal > 0 {
			pct = math.Round(float64(amount)/float64(expenseTotalVal)*1000) / 10
		}
		parentName := ""
		if r.ParentCategoryName.Valid {
			parentName = r.ParentCategoryName.String
		}
		categories = append(categories, ReportCategoryRow{
			CategoryID:         r.CategoryID,
			CategoryName:       r.CategoryName,
			ParentCategoryName: parentName,
			Amount:             amount,
			Percent:            pct,
			TransactionCount:   r.TransactionCount,
		})
	}

	return &ReportResult{
		BaseCurrency:     baseCurrency,
		Period:           filters.Period,
		ExpenseTotal:     expenseTotalVal,
		TransactionCount: txnCount,
		ByCategory:       categories,
	}, nil
}

func (m *ReportManager) generateAccountBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := m.genReportParams(filters)
	ctx := context.Background()

	rows, err := m.q.ReportByAccount(ctx, arg)
	if err != nil {
		return nil, err
	}

	accounts := make([]ReportAccountRow, 0, len(rows))
	for _, r := range rows {
		income := genInterfaceToInt64(r.Income)
		expense := genInterfaceToInt64(r.Expense)
		accounts = append(accounts, ReportAccountRow{
			AccountID:   r.AccountID,
			AccountName: r.AccountName,
			Currency:    r.Currency,
			Income:      income,
			Expense:     expense,
			Net:         income - expense,
		})
	}

	return &ReportResult{
		BaseCurrency: baseCurrency,
		Period:       filters.Period,
		ByAccount:    accounts,
	}, nil
}

func (m *ReportManager) generateTagBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := m.genReportParams(filters)
	ctx := context.Background()

	expenseTotal, err := m.q.ReportExpenseTotal(ctx, gen.ReportExpenseTotalParams(arg))
	if err != nil {
		return nil, err
	}
	expenseTotalVal := genInterfaceToInt64(expenseTotal)

	taggedRows, err := m.q.ReportByTag(ctx, gen.ReportByTagParams(arg))
	if err != nil {
		return nil, err
	}

	untaggedRow, err := m.q.ReportUntagged(ctx, gen.ReportUntaggedParams(arg))
	if err != nil {
		return nil, err
	}

	txnCount, err := m.q.ReportTransactionCount(ctx, gen.ReportTransactionCountParams(arg))
	if err != nil {
		return nil, err
	}

	tags := make([]ReportTagRow, 0, len(taggedRows)+1)

	for _, r := range taggedRows {
		amount := nullFloatToInt64(r.Total)
		pct := float64(0)
		if expenseTotalVal > 0 {
			pct = math.Round(float64(amount)/float64(expenseTotalVal)*1000) / 10
		}
		tags = append(tags, ReportTagRow{
			TagID:            r.TagID,
			TagName:          r.TagName,
			Amount:           amount,
			Percent:          pct,
			TransactionCount: r.TransactionCount,
		})
	}

	if untaggedRow.TransactionCount > 0 {
		untaggedAmount := genInterfaceToInt64(untaggedRow.Total)
		pct := float64(0)
		if expenseTotalVal > 0 {
			pct = math.Round(float64(untaggedAmount)/float64(expenseTotalVal)*1000) / 10
		}
		tags = append(tags, ReportTagRow{
			TagID:            0,
			TagName:          "(untagged)",
			Amount:           untaggedAmount,
			Percent:          pct,
			TransactionCount: untaggedRow.TransactionCount,
		})
	}

	return &ReportResult{
		BaseCurrency:     baseCurrency,
		Period:           filters.Period,
		ExpenseTotal:     expenseTotalVal,
		TransactionCount: txnCount,
		ByTag:            tags,
	}, nil
}
