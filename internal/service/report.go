package service

import (
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type ReportParams struct {
	AccountName  string
	CategoryName string
	Month        string
	DateFrom     string
	DateTo       string
}

type ReportCategoryBreakdown struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Income       int64  `json:"income"`
	Expense      int64  `json:"expense"`
	Net          int64  `json:"net"`
}

type ReportAccountBreakdown struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Currency    string `json:"currency"`
	Income      int64  `json:"income"`
	Expense     int64  `json:"expense"`
	Net         int64  `json:"net"`
}

type ReportResult struct {
	BaseCurrency    string                    `json:"base_currency"`
	TotalIncome     int64                     `json:"total_income"`
	TotalExpense    int64                     `json:"total_expense"`
	Net             int64                     `json:"net"`
	ByCategory      []*ReportCategoryBreakdown `json:"by_category"`
	ByAccount       []*ReportAccountBreakdown  `json:"by_account"`
}

func (s *Service) GenerateReport(params ReportParams) (*ReportResult, error) {
	baseCurrency, err := s.GetBaseCurrency()
	if err != nil {
		return nil, err
	}

	transactions, err := s.fetchReportTransactions(params)
	if err != nil {
		return nil, err
	}

	categories, err := s.ListAllCategories()
	if err != nil {
		return nil, err
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[int64]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	accountMap := make(map[int64]*gen.Account)
	for _, a := range accounts {
		accountMap[a.ID] = a
	}

	catIncome := make(map[int64]int64)
	catExpense := make(map[int64]int64)
	acctIncome := make(map[int64]int64)
	acctExpense := make(map[int64]int64)
	var totalIncome, totalExpense int64

	for _, t := range transactions {
		amount := t.Amount
		if t.BaseAmount.Valid {
			amount = t.BaseAmount.Int64
		}

		if t.Type == "adjustment" {
			continue
		}
		if t.Type == "transfer" {
			continue
		}

		categoryID := int64(0)
		if t.CategoryID.Valid {
			categoryID = t.CategoryID.Int64
		}

		switch t.Type {
		case "income":
			totalIncome += amount
			catIncome[categoryID] += amount
			acctIncome[t.AccountID] += amount
		case "expense":
			totalExpense += amount
			catExpense[categoryID] += amount
			acctExpense[t.AccountID] += amount
		}
	}

	byCategory := make([]*ReportCategoryBreakdown, 0)
	for catID, catName := range categoryMap {
		income := catIncome[catID]
		expense := catExpense[catID]
		if income == 0 && expense == 0 {
			continue
		}
		byCategory = append(byCategory, &ReportCategoryBreakdown{
			CategoryID:   catID,
			CategoryName: catName,
			Income:       income,
			Expense:      expense,
			Net:          income - expense,
		})
	}

	byAccount := make([]*ReportAccountBreakdown, 0)
	for _, a := range accounts {
		income := acctIncome[a.ID]
		expense := acctExpense[a.ID]
		if income == 0 && expense == 0 {
			continue
		}
		byAccount = append(byAccount, &ReportAccountBreakdown{
			AccountID:   a.ID,
			AccountName: a.Name,
			Currency:    a.Currency,
			Income:      income,
			Expense:     expense,
			Net:         income - expense,
		})
	}

	return &ReportResult{
		BaseCurrency: baseCurrency,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Net:          totalIncome - totalExpense,
		ByCategory:   byCategory,
		ByAccount:    byAccount,
	}, nil
}

func (s *Service) fetchReportTransactions(params ReportParams) ([]*gen.Transaction, error) {
	var accountID, categoryID interface{}
	var dateFrom, dateTo interface{}

	if params.AccountName != "" {
		account, err := s.ResolveAccount(params.AccountName)
		if err != nil {
			return nil, err
		}
		accountID = account.ID
	}

	if params.CategoryName != "" {
		category, err := s.ResolveCategory(params.CategoryName)
		if err != nil {
			return nil, err
		}
		categoryID = category.ID
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
			return nil, err
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

	transactions, err := s.q.ListTransactions(s.ctx(), gen.ListTransactionsParams{
		AccountID:  accountID,
		CategoryID: categoryID,
		Type:       nil,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Limit:      10000,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
