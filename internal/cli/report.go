package cli

import (
	"database/sql"
	"fmt"
	"io"
	"sort"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var account, category, month, dateFrom, dateTo string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate financial reports",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runReport(cmd, svc, account, category, month, dateFrom, dateTo)
		}),
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Filter by month (e.g., july, 2026-07)")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Filter by account name or ID")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category name or ID")
	cmd.Flags().StringVar(&dateFrom, "from", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "to", "", "Filter to date (YYYY-MM-DD)")

	return cmd
}

func runReport(cmd *cobra.Command, svc *service.Service, account, category, month, dateFrom, dateTo string) error {
	result, err := svc.GenerateReport(service.ReportParams{
		AccountName:  account,
		CategoryName: category,
		Month:        month,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, result, cmd)
	}

	printReportText(stdout, result)
	return nil
}

func printReportText(stdout io.Writer, result *service.ReportResult) {
	_, _ = fmt.Fprintf(stdout, "Financial Report\n")
	_, _ = fmt.Fprintf(stdout, "=================\n")
	_, _ = fmt.Fprintf(stdout, "Base Currency: %s\n\n", result.BaseCurrency)

	if result.TotalIncome == 0 && result.TotalExpense == 0 {
		_, _ = fmt.Fprintln(stdout, "No transactions found for the selected period.")
		return
	}

	_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "Total Income:", formatAmount(result.TotalIncome))
	_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "Total Expense:", formatAmount(result.TotalExpense))
	if result.Net >= 0 {
		_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "Net:", formatAmount(result.Net))
	} else {
		_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "Net:", formatAmount(result.Net))
	}

	if len(result.ByCategory) > 0 {
		_, _ = fmt.Fprintf(stdout, "\nBy Category\n-----------\n")
		_, _ = fmt.Fprintf(stdout, "%-20s %-8s %-8s %-8s\n", "Category", "Income", "Expense", "Net")

		sort.Slice(result.ByCategory, func(i, j int) bool {
			return result.ByCategory[i].CategoryName < result.ByCategory[j].CategoryName
		})

		for _, cb := range result.ByCategory {
			_, _ = fmt.Fprintf(stdout, "%-20s %-8s %-8s %-8s\n",
				cb.CategoryName,
				formatAmount(cb.Income),
				formatAmount(cb.Expense),
				formatAmount(cb.Net))
		}
	}

	if len(result.ByAccount) > 0 {
		_, _ = fmt.Fprintf(stdout, "\nBy Account\n----------\n")
		_, _ = fmt.Fprintf(stdout, "%-20s %-8s %-8s %-8s\n", "Account", "Income", "Expense", "Net")

		sort.Slice(result.ByAccount, func(i, j int) bool {
			return result.ByAccount[i].AccountName < result.ByAccount[j].AccountName
		})

		for _, ab := range result.ByAccount {
			_, _ = fmt.Fprintf(stdout, "%-20s %-8s %-8s %-8s\n",
				ab.AccountName,
				formatAmount(ab.Income),
				formatAmount(ab.Expense),
				formatAmount(ab.Net))
		}
	}
}
