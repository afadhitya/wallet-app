package service

import (
	"database/sql"
	"errors"
	"log/slog"
	"sort"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type ForecastBalanceResult struct {
	Months            int
	AccountID         int64
	AccountName       string
	MonthlyBalances   []MonthlyBalance
	PlannedPayments   []PlannedPaymentOccurrence
	CategoryBreakdown []CategoryBreakdown
	Warnings          []string
}

type MonthlyBalance struct {
	Year              int
	Month             time.Month
	StartBalance      int64
	ProjectedIncome   int64
	ProjectedExpenses int64
	NetMovement       int64
	EndingBalance     int64
	IsNegative        bool
}

type PlannedPaymentOccurrence struct {
	PlannedPaymentID   int64
	PlannedPaymentName string
	DueDate            string
	Amount             int64
	Type               string
	Currency           string
	AccountName        string
	CategoryName       string
}

type CategoryBreakdown struct {
	CategoryName string
	TotalExpense int64
}

type ForecastBillsResult struct {
	Months      int
	Bills       []BillRow
	TotalAmount int64
	Warnings    []string
}

type BillRow struct {
	DueDate      string
	Name         string
	Amount       int64
	RunningTotal int64
}

var todayFunc = time.Now

func (s *Service) ForecastBalance(months int, accountFilter string) (*ForecastBalanceResult, error) {
	s.logger.Info("ForecastBalance called", slog.Int("months", months), slog.String("account_filter", accountFilter))

	if months <= 0 {
		s.logger.Warn("ForecastBalance validation failed", slog.Int("months", months), slog.String("error", "forecast horizon must be positive"))
		return nil, &ValidationError{Field: "months", Message: "forecast horizon must be positive"}
	}

	var resolvedAccount *gen.Account
	var accountID sql.NullInt64
	if accountFilter != "" {
		account, err := s.ResolveAccount(accountFilter)
		if err != nil {
			var nfErr *NotFoundError
			var vErr *ValidationError
			if errors.As(err, &nfErr) || errors.As(err, &vErr) {
				s.logger.Warn("ForecastBalance account resolution failed", slog.String("error", err.Error()))
			} else {
				s.logger.Error("ForecastBalance failed", slog.String("error", err.Error()))
			}
			return nil, err
		}
		resolvedAccount = account
		accountID = sql.NullInt64{Int64: account.ID, Valid: true}
	}

	payments, err := s.q.ListActivePlannedPaymentsForAccount(s.ctx(), accountID)
	if err != nil {
		s.logger.Error("ForecastBalance failed", slog.String("error", err.Error()))
		return nil, err
	}

	accounts := make(map[int64]*gen.Account)
	if resolvedAccount != nil {
		accounts[resolvedAccount.ID] = resolvedAccount
	} else {
		allAccounts, err := s.q.ListAccounts(s.ctx())
		if err != nil {
			s.logger.Error("ForecastBalance failed", slog.String("error", err.Error()))
			return nil, err
		}
		for _, a := range allAccounts {
			accounts[a.ID] = a
		}
	}

	categories := make(map[int64]string)

	today := todayFunc()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	horizonStart := today
	horizonEnd := today.AddDate(0, months, 0)

	type monthOccurrence struct {
		monthIndex int
		occ        PlannedPaymentOccurrence
	}
	var allOccurrences []monthOccurrence

	dateFormat := "2006-01-02"
	for _, pp := range payments {
		if !pp.NextDueDate.Valid {
			continue
		}

		categoryName := "Uncategorized"
		if pp.CategoryID.Valid {
			if name, ok := categories[pp.CategoryID.Int64]; ok {
				categoryName = name
			} else {
				cat, err := s.q.GetCategoryByID(s.ctx(), pp.CategoryID.Int64)
				if err == nil {
					categories[pp.CategoryID.Int64] = cat.Name
					categoryName = cat.Name
				}
			}
		}

		accountName := "Unknown"
		if acc, ok := accounts[pp.AccountID]; ok {
			accountName = acc.Name
		}

		seedDate, err := time.Parse(dateFormat, pp.NextDueDate.String)
		if err != nil {
			continue
		}

		occurrences := expandOccurrences(seedDate, pp.Recurrence, pp.RecurrenceRule, horizonStart, horizonEnd)

		for _, occDate := range occurrences {
			if occDate.Before(horizonStart) || !occDate.Before(horizonEnd) {
				continue
			}
			monthIdx := monthsBetween(horizonStart, occDate)
			if monthIdx < 0 || monthIdx >= months {
				continue
			}
			allOccurrences = append(allOccurrences, monthOccurrence{
				monthIndex: monthIdx,
				occ: PlannedPaymentOccurrence{
					PlannedPaymentID:   pp.ID,
					PlannedPaymentName: pp.Name,
					DueDate:            occDate.Format(dateFormat),
					Amount:             pp.Amount,
					Type:               pp.Type,
					Currency:           pp.Currency,
					AccountName:        accountName,
					CategoryName:       categoryName,
				},
			})
		}
	}

	monthlyBalances := make([]MonthlyBalance, months)
	for i := 0; i < months; i++ {
		monthStart := horizonStart.AddDate(0, i, 0)
		monthlyBalances[i] = MonthlyBalance{
			Year:  monthStart.Year(),
			Month: monthStart.Month(),
		}
	}

	var warnings []string
	missingRateWarned := make(map[string]bool)

	for i := 0; i < months; i++ {
		var startBalance int64
		if accountFilter != "" && resolvedAccount != nil {
			converted, err := s.Convert(resolvedAccount.Balance, resolvedAccount.Currency)
			if err != nil {
				converted = resolvedAccount.Balance
			}
			startBalance = converted
		} else {
			for _, acc := range accounts {
				converted, err := s.Convert(acc.Balance, acc.Currency)
				if err != nil {
					converted = acc.Balance
				}
				startBalance += converted
			}
		}
		if i > 0 {
			startBalance = monthlyBalances[i-1].EndingBalance
		}
		monthlyBalances[i].StartBalance = startBalance

		var income, expenses int64
		for _, mo := range allOccurrences {
			if mo.monthIndex == i {
				convertedAmt, err := s.Convert(mo.occ.Amount, mo.occ.Currency)
				if err != nil {
					warnKey := mo.occ.PlannedPaymentName + ":" + mo.occ.Currency
					if !missingRateWarned[warnKey] {
						missingRateWarned[warnKey] = true
						warnings = append(warnings, "Skipped planned payment \""+mo.occ.PlannedPaymentName+"\": missing exchange rate for "+mo.occ.Currency)
					}
					continue
				}
				if mo.occ.Type == "income" {
					income += convertedAmt
				} else {
					expenses += convertedAmt
				}
			}
		}
		monthlyBalances[i].ProjectedIncome = income
		monthlyBalances[i].ProjectedExpenses = expenses
		monthlyBalances[i].NetMovement = income - expenses
		monthlyBalances[i].EndingBalance = startBalance + income - expenses
		if monthlyBalances[i].EndingBalance < 0 {
			monthlyBalances[i].IsNegative = true
		}
	}

	for _, mb := range monthlyBalances {
		if mb.IsNegative {
			monthLabel := time.Date(mb.Year, mb.Month, 1, 0, 0, 0, 0, time.UTC).Format("January 2006")
			if accountFilter != "" && resolvedAccount != nil {
				warnings = append(warnings, "Projected negative balance for "+resolvedAccount.Name+" in "+monthLabel)
			} else {
				warnings = append(warnings, "Projected negative balance in "+monthLabel)
			}
		}
	}

	var plannedPaymentList []PlannedPaymentOccurrence
	for _, mo := range allOccurrences {
		plannedPaymentList = append(plannedPaymentList, mo.occ)
	}

	categoryExpenses := make(map[string]int64)
	for _, mo := range allOccurrences {
		if mo.occ.Type == "expense" {
			convertedAmt, err := s.Convert(mo.occ.Amount, mo.occ.Currency)
			if err != nil {
				convertedAmt = mo.occ.Amount
			}
			categoryExpenses[mo.occ.CategoryName] += convertedAmt
		}
	}
	var categoryBreakdown []CategoryBreakdown
	for cat, total := range categoryExpenses {
		categoryBreakdown = append(categoryBreakdown, CategoryBreakdown{
			CategoryName: cat,
			TotalExpense: total,
		})
	}
	sort.Slice(categoryBreakdown, func(i, j int) bool {
		return categoryBreakdown[i].TotalExpense > categoryBreakdown[j].TotalExpense
	})

	result := &ForecastBalanceResult{
		Months:            months,
		MonthlyBalances:   monthlyBalances,
		PlannedPayments:   plannedPaymentList,
		CategoryBreakdown: categoryBreakdown,
		Warnings:          warnings,
	}

	if resolvedAccount != nil {
		result.AccountID = resolvedAccount.ID
		result.AccountName = resolvedAccount.Name
	}

	s.logger.Info("ForecastBalance completed",
		slog.Int("months", months),
		slog.Int("num_warnings", len(warnings)),
		slog.Int("num_payments", len(plannedPaymentList)),
		slog.Int("num_categories", len(categoryBreakdown)),
	)
	return result, nil
}

func (s *Service) ForecastBills(months int) (*ForecastBillsResult, error) {
	s.logger.Info("ForecastBills called", slog.Int("months", months))

	if months <= 0 {
		s.logger.Warn("ForecastBills validation failed", slog.Int("months", months), slog.String("error", "forecast horizon must be positive"))
		return nil, &ValidationError{Field: "months", Message: "forecast horizon must be positive"}
	}

	payments, err := s.q.ListActivePlannedPaymentsForAccount(s.ctx(), sql.NullInt64{})
	if err != nil {
		s.logger.Error("ForecastBills failed", slog.String("error", err.Error()))
		return nil, err
	}

	today := todayFunc()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	horizonStart := today
	horizonEnd := today.AddDate(0, months, 0)

	dateFormat := "2006-01-02"

	type billOccurrence struct {
		Date    time.Time
		BillRow BillRow
	}
	var bills []billOccurrence

	for _, pp := range payments {
		if pp.Type != "expense" {
			continue
		}
		if !pp.NextDueDate.Valid {
			continue
		}

		seedDate, err := time.Parse(dateFormat, pp.NextDueDate.String)
		if err != nil {
			continue
		}

		occurrences := expandOccurrences(seedDate, pp.Recurrence, pp.RecurrenceRule, horizonStart, horizonEnd)

		for _, occDate := range occurrences {
			if occDate.Before(horizonStart) || !occDate.Before(horizonEnd) {
				continue
			}
			bills = append(bills, billOccurrence{
				Date: occDate,
				BillRow: BillRow{
					DueDate: occDate.Format(dateFormat),
					Name:    pp.Name,
					Amount:  pp.Amount,
				},
			})
		}
	}

	sort.Slice(bills, func(i, j int) bool {
		return bills[i].Date.Before(bills[j].Date)
	})

	var runningTotal int64
	var billRows []BillRow
	for i := range bills {
		runningTotal += bills[i].BillRow.Amount
		bills[i].BillRow.RunningTotal = runningTotal
		billRows = append(billRows, bills[i].BillRow)
	}

	var warnings []string
	if len(billRows) == 0 {
		warnings = append(warnings, "No planned payments found. Forecasts are based on planned payments only.")
	}

	s.logger.Info("ForecastBills completed",
		slog.Int("months", months),
		slog.Int("num_bills", len(billRows)),
		slog.Int64("total_amount", runningTotal),
	)
	return &ForecastBillsResult{
		Months:      months,
		Bills:       billRows,
		TotalAmount: runningTotal,
		Warnings:    warnings,
	}, nil
}

func expandOccurrences(seedDate time.Time, recurrence string, recurrenceRule sql.NullString, horizonStart, horizonEnd time.Time) []time.Time {
	if recurrence == "none" {
		if !seedDate.Before(horizonStart) && seedDate.Before(horizonEnd) {
			return []time.Time{seedDate}
		}
		return nil
	}

	var occurrences []time.Time
	current := seedDate

	for !current.After(horizonEnd) && !current.Equal(horizonEnd) {
		if !current.Before(horizonStart) {
			occurrences = append(occurrences, current)
		}
		next, err := calcNextDue(current, recurrence, recurrenceRule)
		if err != nil {
			break
		}
		if !next.After(current) {
			break
		}
		current = next
	}

	return occurrences
}

func monthsBetween(start, date time.Time) int {
	years := date.Year() - start.Year()
	months := int(date.Month()) - int(start.Month())
	result := years*12 + months
	if date.Day() < start.Day() {
		result--
	}
	return result
}
