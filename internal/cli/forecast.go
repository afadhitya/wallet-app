package cli

import (
	"database/sql"
	"fmt"
	"io"
	"sort"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newForecastCmd() *cobra.Command {
	var months int
	var account string

	cmd := &cobra.Command{
		Use:   "forecast",
		Short: "Forecast future balances and bills from planned payments",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runForecastBalance(cmd, svc, months, account)
		}),
	}

	cmd.Flags().IntVarP(&months, "months", "n", 1, "Forecast horizon in months (default: 1)")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Limit forecast to a specific account")

	cmd.AddCommand(newForecastBillsCmd())

	return cmd
}

func newForecastBillsCmd() *cobra.Command {
	var months int

	cmd := &cobra.Command{
		Use:   "bills",
		Short: "Show upcoming bill impact from planned payments",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runForecastBills(cmd, svc, months)
		}),
	}

	cmd.Flags().IntVarP(&months, "months", "n", 2, "Forecast horizon in months (default: 2)")

	return cmd
}

func runForecastBalance(cmd *cobra.Command, svc *service.Service, months int, account string) error {
	result, err := svc.ForecastBalance(months, account)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printBalanceForecastJSON(stdout, result)
	}
	return printBalanceForecastText(stdout, result)
}

func runForecastBills(cmd *cobra.Command, svc *service.Service, months int) error {
	result, err := svc.ForecastBills(months)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printBillsForecastJSON(stdout, result)
	}
	return printBillsForecastText(stdout, result)
}

func printBalanceForecastText(w io.Writer, result *service.ForecastBalanceResult) error {
	if len(result.PlannedPayments) == 0 {
		_, _ = fmt.Fprintln(w, "No planned payments found. Forecasts are based on planned payments only.")
		return nil
	}

	scopeLabel := "All Accounts"
	if result.AccountName != "" {
		scopeLabel = result.AccountName
	}
	_, _ = fmt.Fprintf(w, "Forecast: %s (%d month", scopeLabel, result.Months)
	if result.Months > 1 {
		_, _ = fmt.Fprint(w, "s")
	}
	_, _ = fmt.Fprintln(w, ")")
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "%-12s %14s %14s %14s %14s %14s\n",
		"Month", "Start Balance", "Income", "Expenses", "Net Movement", "Ending Balance")
	_, _ = fmt.Fprintf(w, "%-12s %14s %14s %14s %14s %14s\n",
		"------------", "-------------", "-------------", "-------------", "-------------", "-------------")

	for _, mb := range result.MonthlyBalances {
		monthLabel := fmt.Sprintf("%s %d", mb.Month.String()[:3], mb.Year)
		endLabel := formatAmount(mb.EndingBalance)
		if mb.IsNegative {
			endLabel += " *"
		}
		_, _ = fmt.Fprintf(w, "%-12s %14s %14s %14s %14s %14s\n",
			monthLabel,
			formatAmount(mb.StartBalance),
			formatAmount(mb.ProjectedIncome),
			"-"+formatAmount(mb.ProjectedExpenses),
			formatAmount(mb.NetMovement),
			endLabel,
		)
	}

	if len(result.PlannedPayments) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "%-22s %-12s %-10s %-10s\n", "Name", "Due Date", "Amount", "Type")
		_, _ = fmt.Fprintf(w, "%-22s %-12s %-10s %-10s\n", "----------------------", "------------", "----------", "----------")

		sort.Slice(result.PlannedPayments, func(i, j int) bool {
			return result.PlannedPayments[i].DueDate < result.PlannedPayments[j].DueDate
		})

		for _, pp := range result.PlannedPayments {
			_, _ = fmt.Fprintf(w, "%-22s %-12s %-10s %-10s\n",
				truncate(pp.PlannedPaymentName, 22),
				pp.DueDate,
				formatAmount(pp.Amount),
				pp.Type,
			)
		}
	}

	if len(result.CategoryBreakdown) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "Category Breakdown:")
		for _, cb := range result.CategoryBreakdown {
			_, _ = fmt.Fprintf(w, "  %-20s %s\n", cb.CategoryName, formatAmount(cb.TotalExpense))
		}
	}

	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintln(w)
		for _, warn := range result.Warnings {
			_, _ = fmt.Fprintf(w, "Warning: %s\n", warn)
		}
	}

	return nil
}

func printBillsForecastText(w io.Writer, result *service.ForecastBillsResult) error {
	if len(result.Bills) == 0 {
		_, _ = fmt.Fprintln(w, "No planned payments found. Forecasts are based on planned payments only.")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Upcoming Bills (%d month", result.Months)
	if result.Months > 1 {
		_, _ = fmt.Fprint(w, "s")
	}
	_, _ = fmt.Fprintln(w, ")")
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "%-12s %-22s %-14s %14s\n", "Due Date", "Name", "Amount", "Running Total")
	_, _ = fmt.Fprintf(w, "%-12s %-22s %-14s %14s\n", "------------", "----------------------", "--------------", "-------------")

	for _, bill := range result.Bills {
		_, _ = fmt.Fprintf(w, "%-12s %-22s %-14s %14s\n",
			bill.DueDate,
			truncate(bill.Name, 22),
			formatAmount(bill.Amount),
			formatAmount(bill.RunningTotal),
		)
	}

	if result.TotalAmount > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "Total: %s across %d bill(s)\n", formatAmount(result.TotalAmount), len(result.Bills))
	}

	return nil
}

func printBalanceForecastJSON(w io.Writer, result *service.ForecastBalanceResult) error {
	type jsonMonthlyBalance struct {
		Year              int    `json:"year"`
		Month             string `json:"month"`
		StartBalance      int64  `json:"start_balance"`
		ProjectedIncome   int64  `json:"projected_income"`
		ProjectedExpenses int64  `json:"projected_expenses"`
		NetMovement       int64  `json:"net_movement"`
		EndingBalance     int64  `json:"ending_balance"`
		IsNegative        bool   `json:"is_negative"`
	}

	type jsonPlannedPaymentOccurrence struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		DueDate      string `json:"due_date"`
		Amount       int64  `json:"amount"`
		Type         string `json:"type"`
		Currency     string `json:"currency"`
		AccountName  string `json:"account_name"`
		CategoryName string `json:"category_name"`
	}

	type jsonCategoryBreakdown struct {
		CategoryName string `json:"category_name"`
		TotalExpense int64  `json:"total_expense"`
	}

	var mbs []jsonMonthlyBalance
	for _, mb := range result.MonthlyBalances {
		mbs = append(mbs, jsonMonthlyBalance{
			Year:              mb.Year,
			Month:             mb.Month.String(),
			StartBalance:      mb.StartBalance,
			ProjectedIncome:   mb.ProjectedIncome,
			ProjectedExpenses: mb.ProjectedExpenses,
			NetMovement:       mb.NetMovement,
			EndingBalance:     mb.EndingBalance,
			IsNegative:        mb.IsNegative,
		})
	}

	var pps []jsonPlannedPaymentOccurrence
	for _, pp := range result.PlannedPayments {
		pps = append(pps, jsonPlannedPaymentOccurrence{
			ID:           pp.PlannedPaymentID,
			Name:         pp.PlannedPaymentName,
			DueDate:      pp.DueDate,
			Amount:       pp.Amount,
			Type:         pp.Type,
			Currency:     pp.Currency,
			AccountName:  pp.AccountName,
			CategoryName: pp.CategoryName,
		})
	}

	var cbs []jsonCategoryBreakdown
	for _, cb := range result.CategoryBreakdown {
		cbs = append(cbs, jsonCategoryBreakdown{
			CategoryName: cb.CategoryName,
			TotalExpense: cb.TotalExpense,
		})
	}

	jsonResponse := map[string]interface{}{
		"horizon":             result.Months,
		"forecast":            mbs,
		"planned_payments":    pps,
		"category_breakdown":  cbs,
		"warnings":            result.Warnings,
	}

	if result.AccountName != "" {
		jsonResponse["account"] = map[string]interface{}{
			"id":   result.AccountID,
			"name": result.AccountName,
		}
	}

	return printJSON(w, jsonResponse)
}

func printBillsForecastJSON(w io.Writer, result *service.ForecastBillsResult) error {
	type jsonBillRow struct {
		DueDate      string `json:"due_date"`
		Name         string `json:"name"`
		Amount       int64  `json:"amount"`
		RunningTotal int64  `json:"running_total"`
	}

	var brs []jsonBillRow
	for _, b := range result.Bills {
		brs = append(brs, jsonBillRow{
			DueDate:      b.DueDate,
			Name:         b.Name,
			Amount:       b.Amount,
			RunningTotal: b.RunningTotal,
		})
	}

	jsonResponse := map[string]interface{}{
		"horizon":      result.Months,
		"bills":        brs,
		"total_amount": result.TotalAmount,
		"count":        len(result.Bills),
		"warnings":     result.Warnings,
	}

	return printJSON(w, jsonResponse)
}
