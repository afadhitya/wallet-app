package service

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type ReportParams struct {
	Month       string
	DateFrom    string
	DateTo      string
	AccountName string
	By          string
	Export      string
	OutputPath  string
}

type ReportFilters struct {
	DateFrom  string
	DateTo    string
	Period    string
	AccountID *int64
}

type ReportCategoryRow struct {
	CategoryID         int64   `json:"category_id"`
	CategoryName       string  `json:"category_name"`
	ParentCategoryName string  `json:"parent_category_name,omitempty"`
	Amount             int64   `json:"amount"`
	Percent            float64 `json:"percent"`
	TransactionCount   int64   `json:"transaction_count"`
}

type ReportAccountRow struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Currency    string `json:"currency"`
	Income      int64  `json:"income"`
	Expense     int64  `json:"expense"`
	Net         int64  `json:"net"`
}

type ReportTagRow struct {
	TagID            int64   `json:"tag_id"`
	TagName          string  `json:"tag_name"`
	Amount           int64   `json:"amount"`
	Percent          float64 `json:"percent"`
	TransactionCount int64   `json:"transaction_count"`
}

type ReportExportRow struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	BaseAmount  int64  `json:"base_amount"`
	Category    string `json:"category"`
	Account     string `json:"account"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

type ReportResult struct {
	BaseCurrency     string              `json:"base_currency"`
	Period           string              `json:"period"`
	IncomeTotal      int64               `json:"income_total"`
	ExpenseTotal     int64               `json:"expense_total"`
	Net              int64               `json:"net"`
	TransferTotal    int64               `json:"transfer_total"`
	TransactionCount int64               `json:"transaction_count"`
	IncomeCategories []ReportCategoryRow `json:"income_categories,omitempty"`
	ExpenseCategories []ReportCategoryRow `json:"expense_categories,omitempty"`
	ByCategory       []ReportCategoryRow `json:"by_category,omitempty"`
	ByAccount        []ReportAccountRow  `json:"by_account,omitempty"`
	ByTag            []ReportTagRow      `json:"by_tag,omitempty"`
	ExportRows       []ReportExportRow   `json:"export_rows,omitempty"`
}

var (
	ErrInvalidMonth  = &ValidationError{Field: "month", Message: "Invalid month format. Expected month name or YYYY-MM."}
	ErrInvalidExport = &ValidationError{Field: "export", Message: "Unsupported export format. Only 'csv' is supported."}
	ErrInvalidBy     = &ValidationError{Field: "by", Message: "Unsupported breakdown. Expected 'category', 'account', or 'tag'."}
	ErrExportFailed  = fmt.Errorf("export failed")
	ErrNoReportData  = fmt.Errorf("no transactions found for specified period")
)

func (s *Service) GenerateReport(params ReportParams) (*ReportResult, error) {
	baseCurrency, err := s.GetBaseCurrency()
	if err != nil {
		return nil, err
	}

	filters, err := s.resolveReportFilters(params)
	if err != nil {
		return nil, err
	}

	switch params.By {
	case "":
		return s.generateMonthlySummary(baseCurrency, filters)
	case "category":
		return s.generateCategoryBreakdownResult(baseCurrency, filters)
	case "account":
		return s.generateAccountBreakdownResult(baseCurrency, filters)
	case "tag":
		return s.generateTagBreakdownResult(baseCurrency, filters)
	default:
		return nil, ErrInvalidBy
	}
}

func (s *Service) resolveReportFilters(params ReportParams) (ReportFilters, error) {
	var filters ReportFilters

	period, dateFrom, dateTo, err := s.resolvePeriod(params)
	if err != nil {
		return filters, err
	}
	filters.Period = period
	filters.DateFrom = dateFrom
	filters.DateTo = dateTo

	if params.AccountName != "" {
		account, err := s.ResolveAccount(params.AccountName)
		if err != nil {
			return filters, err
		}
		filters.AccountID = &account.ID
	}

	return filters, nil
}

func (s *Service) resolvePeriod(params ReportParams) (string, string, string, error) {
	now := time.Now()

	if params.DateFrom != "" || params.DateTo != "" {
		if params.DateFrom == "" {
			return "", "", "", &ValidationError{Field: "from", Message: "date range requires both --from and --to"}
		}
		if params.DateTo == "" {
			return "", "", "", &ValidationError{Field: "to", Message: "date range requires both --from and --to"}
		}
		if _, err := time.Parse("2006-01-02", params.DateFrom); err != nil {
			return "", "", "", &ValidationError{Field: "from", Message: "date must be YYYY-MM-DD"}
		}
		if _, err := time.Parse("2006-01-02", params.DateTo); err != nil {
			return "", "", "", &ValidationError{Field: "to", Message: "date must be YYYY-MM-DD"}
		}
		period := params.DateFrom + " to " + params.DateTo
		return period, params.DateFrom, params.DateTo, nil
	}

	if params.Month != "" {
		from, to, err := parseMonth(params.Month)
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

func (s *Service) genReportParams(filters ReportFilters) gen.ReportByAccountParams {
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

func (s *Service) generateMonthlySummary(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := s.genReportParams(filters)

	incomeTotal, err := s.q.ReportIncomeTotal(s.ctx(), gen.ReportIncomeTotalParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	expenseTotal, err := s.q.ReportExpenseTotal(s.ctx(), gen.ReportExpenseTotalParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	transfers, err := s.q.ReportTransfers(s.ctx(), gen.ReportTransfersParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	txnCount, err := s.q.ReportTransactionCount(s.ctx(), gen.ReportTransactionCountParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	incomeTotalVal := genInterfaceToInt64(incomeTotal)
	expenseTotalVal := genInterfaceToInt64(expenseTotal)
	transferTotal := genInterfaceToInt64(transfers)

	incomeRows, err := s.q.ReportIncomeByCategory(s.ctx(), gen.ReportIncomeByCategoryParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	expenseRows, err := s.q.ReportExpenseByCategory(s.ctx(), gen.ReportExpenseByCategoryParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
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

func (s *Service) generateCategoryBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := s.genReportParams(filters)

	expenseTotal, err := s.q.ReportExpenseTotal(s.ctx(), gen.ReportExpenseTotalParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}
	expenseTotalVal := genInterfaceToInt64(expenseTotal)

	rows, err := s.q.ReportExpenseByCategory(s.ctx(), gen.ReportExpenseByCategoryParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	txnCount, err := s.q.ReportTransactionCount(s.ctx(), gen.ReportTransactionCountParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
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

func (s *Service) generateAccountBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := s.genReportParams(filters)

	rows, err := s.q.ReportByAccount(s.ctx(), arg)
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

func (s *Service) generateTagBreakdownResult(baseCurrency string, filters ReportFilters) (*ReportResult, error) {
	arg := s.genReportParams(filters)

	expenseTotal, err := s.q.ReportExpenseTotal(s.ctx(), gen.ReportExpenseTotalParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}
	expenseTotalVal := genInterfaceToInt64(expenseTotal)

	taggedRows, err := s.q.ReportByTag(s.ctx(), gen.ReportByTagParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	untaggedRow, err := s.q.ReportUntagged(s.ctx(), gen.ReportUntaggedParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	txnCount, err := s.q.ReportTransactionCount(s.ctx(), gen.ReportTransactionCountParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
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

func (s *Service) GenerateExportRows(params ReportParams) ([]ReportExportRow, error) {
	filters, err := s.resolveReportFilters(params)
	if err != nil {
		return nil, err
	}

	arg := s.genReportParams(filters)

	rows, err := s.q.ReportExportTransactions(s.ctx(), gen.ReportExportTransactionsParams{
		DateFrom:  arg.DateFrom,
		DateTo:    arg.DateTo,
		AccountID: arg.AccountID,
	})
	if err != nil {
		return nil, err
	}

	exportRows := make([]ReportExportRow, 0, len(rows))
	for _, r := range rows {
		baseAmount := int64(0)
		if r.BaseAmount.Valid {
			baseAmount = r.BaseAmount.Int64
		}

		categoryName := ""
		if r.CategoryName.Valid {
			categoryName = r.CategoryName.String
		}

		description := ""
		if r.Description.Valid {
			description = r.Description.String
		}

		tags, err := s.q.ListTransactionTags(s.ctx(), r.ID)
		if err != nil {
			return nil, err
		}

		tagNames := make([]string, 0, len(tags))
		for _, t := range tags {
			tagNames = append(tagNames, t.Name)
		}

		exportRows = append(exportRows, ReportExportRow{
			Date:        r.Date,
			Type:        r.Type,
			Amount:      r.Amount,
			Currency:    r.Currency,
			BaseAmount:  baseAmount,
			Category:    categoryName,
			Account:     r.AccountName,
			Description: description,
			Tags:        joinTagNames(tagNames),
		})
	}

	return exportRows, nil
}

func joinTagNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += "," + names[i]
	}
	return result
}

func (s *Service) DefaultExportFilename(params ReportParams) (string, error) {
	filters, err := s.resolveReportFilters(params)
	if err != nil {
		return "", err
	}

	t, err := time.Parse("2006-01-02", filters.DateFrom)
	if err != nil {
		return fmt.Sprintf("wallet-report-%s.csv", filters.DateFrom), nil
	}

	return fmt.Sprintf("wallet-report-%s.csv", t.Format("2006-01")), nil
}
