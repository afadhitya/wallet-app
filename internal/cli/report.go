package cli

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var account, month, dateFrom, dateTo, by, export, output string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate financial reports",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			params := service.ReportParams{
				Month:       month,
				DateFrom:    dateFrom,
				DateTo:      dateTo,
				AccountName: account,
				By:          by,
				Export:      export,
				OutputPath:  output,
			}

			if by != "" {
				switch by {
				case "category", "account", "tag":
				default:
					if isJSON(cmd) {
						_, stderr := resolveOut(cmd)
						_ = printErrorJSON(stderr, "INVALID_INPUT", "Unsupported breakdown type. Expected 'category', 'account', or 'tag'.", "")
					}
					return formatError(cmd, service.ErrInvalidBy)
				}
			}

			if export != "" {
				if export != "csv" {
					if isJSON(cmd) {
						_, stderr := resolveOut(cmd)
						_ = printErrorJSON(stderr, "INVALID_INPUT", "Unsupported export format. Only 'csv' is supported.", "")
					}
					return formatError(cmd, service.ErrInvalidExport)
				}
				return runReportExport(cmd, svc, params)
			}

			return runReport(cmd, svc, params)
		}),
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Filter by month (e.g., july, 2026-07)")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Filter by account name or ID")
	cmd.Flags().StringVar(&dateFrom, "from", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "to", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&by, "by", "", "Breakdown type: category, account, or tag")
	cmd.Flags().StringVar(&export, "export", "", "Export format (csv)")
	cmd.Flags().StringVar(&output, "output", "", "Custom output filename for CSV export")

	return cmd
}

func runReport(cmd *cobra.Command, svc *service.Service, params service.ReportParams) error {
	result, err := svc.GenerateReport(params)
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

func runReportExport(cmd *cobra.Command, svc *service.Service, params service.ReportParams) error {
	rows, err := svc.GenerateExportRows(params)
	if err != nil {
		return formatError(cmd, err)
	}

	outputPath := params.OutputPath
	if outputPath == "" {
		outputPath, err = svc.DefaultExportFilename(params)
		if err != nil {
			return formatError(cmd, err)
		}
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		result := struct {
			FilePath string                  `json:"file_path"`
			Format   string                  `json:"format"`
			Rows     []service.ReportExportRow `json:"rows"`
		}{
			FilePath: outputPath,
			Format:   "csv",
			Rows:     rows,
		}
		return printSuccessJSON(stdout, result, cmd)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return formatError(cmd, fmt.Errorf("failed to export: %w", err))
	}
	defer func() { _ = file.Close() }()

	writer := csv.NewWriter(file)
	header := []string{"date", "type", "amount", "currency", "base_amount", "category", "account", "description", "tags"}
	if err := writer.Write(header); err != nil {
		return formatError(cmd, fmt.Errorf("failed to export: %w", err))
	}

	for _, row := range rows {
		baseAmountStr := ""
		if row.BaseAmount > 0 {
			baseAmountStr = fmt.Sprintf("%d", row.BaseAmount)
		}
		record := []string{
			row.Date,
			row.Type,
			fmt.Sprintf("%d", row.Amount),
			row.Currency,
			baseAmountStr,
			row.Category,
			row.Account,
			row.Description,
			row.Tags,
		}
		if err := writer.Write(record); err != nil {
			return formatError(cmd, fmt.Errorf("failed to export: %w", err))
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return formatError(cmd, fmt.Errorf("failed to export: %w", err))
	}

	_, _ = fmt.Fprintf(stdout, "Exported to: %s\n", outputPath)
	return nil
}

func printReportText(stdout io.Writer, result *service.ReportResult) {
	if result.IncomeTotal == 0 && result.ExpenseTotal == 0 && len(result.ByAccount) == 0 && len(result.ByCategory) == 0 && len(result.ByTag) == 0 {
		_, _ = fmt.Fprintln(stdout, "No transactions found for the specified period.")
		return
	}

	_, _ = fmt.Fprintf(stdout, "Report — %s\n", result.Period)
	_, _ = fmt.Fprintf(stdout, "Base Currency: %s\n\n", result.BaseCurrency)

	if len(result.ByCategory) > 0 {
		printCategoryBreakdown(stdout, result)
	} else if len(result.ByAccount) > 0 {
		printAccountBreakdown(stdout, result)
	} else if len(result.ByTag) > 0 {
		printTagBreakdown(stdout, result)
	} else {
		printMonthlySummary(stdout, result)
	}
}

func printMonthlySummary(stdout io.Writer, result *service.ReportResult) {
	_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "Income:", formatAmount(result.IncomeTotal))
	for _, cat := range result.IncomeCategories {
		_, _ = fmt.Fprintf(stdout, "  %-18s %s\n", cat.CategoryName+":", formatAmount(cat.Amount))
	}

	_, _ = fmt.Fprintf(stdout, "\n%-20s %s\n", "Expenses:", formatAmount(result.ExpenseTotal))
	for _, cat := range result.ExpenseCategories {
		_, _ = fmt.Fprintf(stdout, "  %-18s %s\n", cat.CategoryName+":", formatAmount(cat.Amount))
	}

	netStr := formatAmount(result.Net)
	if result.Net >= 0 {
		netStr = "+" + netStr
	}
	_, _ = fmt.Fprintf(stdout, "\n%-20s %s\n", "Net:", netStr)

	if result.TransferTotal > 0 {
		_, _ = fmt.Fprintf(stdout, "\n%-20s %s (between own accounts)\n", "Transfers:", formatAmount(result.TransferTotal))
	}
}

func printCategoryBreakdown(stdout io.Writer, result *service.ReportResult) {
	_, _ = fmt.Fprintf(stdout, "Expenses by Category\n")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "Category", "Amount", "%", "Transactions")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "--------------------", "----------", "-----", "------------")

	for _, cat := range result.ByCategory {
		prefix := ""
		if cat.ParentCategoryName != "" {
			prefix = "  "
		}
		_, _ = fmt.Fprintf(stdout, "%-20s %s%10s %5.1f%% %12d\n",
			cat.CategoryName,
			prefix,
			formatAmount(cat.Amount),
			cat.Percent,
			cat.TransactionCount,
		)
	}

	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "--------------------", "----------", "-----", "------------")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %5.1f%% %12d\n", "TOTAL", formatAmount(result.ExpenseTotal), 100.0, result.TransactionCount)
}

func printAccountBreakdown(stdout io.Writer, result *service.ReportResult) {
	_, _ = fmt.Fprintf(stdout, "Transactions by Account\n")
	_, _ = fmt.Fprintf(stdout, "%-20s %12s %12s %12s\n", "Account", "Income", "Expenses", "Net")
	_, _ = fmt.Fprintf(stdout, "%-20s %12s %12s %12s\n", "--------------------", "------------", "------------", "------------")

	var totalIncome, totalExpense, totalNet int64
	for _, a := range result.ByAccount {
		netStr := formatAmount(a.Net)
		if a.Net >= 0 {
			netStr = "+" + netStr
		} else {
			netStr = "-" + formatAmount(-a.Net)
		}
		_, _ = fmt.Fprintf(stdout, "%-20s %12s %12s %12s\n",
			a.AccountName,
			formatAmount(a.Income),
			formatAmount(a.Expense),
			netStr,
		)
		totalIncome += a.Income
		totalExpense += a.Expense
		totalNet += a.Net
	}

	totalNetStr := formatAmount(totalNet)
	if totalNet >= 0 {
		totalNetStr = "+" + totalNetStr
	}
	_, _ = fmt.Fprintf(stdout, "%-20s %12s %12s %12s\n", "--------------------", "------------", "------------", "------------")
	_, _ = fmt.Fprintf(stdout, "%-20s %12s %12s %12s\n", "TOTAL", formatAmount(totalIncome), formatAmount(totalExpense), totalNetStr)
}

func printTagBreakdown(stdout io.Writer, result *service.ReportResult) {
	_, _ = fmt.Fprintf(stdout, "Expenses by Tag\n")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "Tag", "Amount", "%", "Transactions")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "--------------------", "----------", "-----", "------------")

	for _, t := range result.ByTag {
		_, _ = fmt.Fprintf(stdout, "%-20s %10s %5.1f%% %12d\n",
			t.TagName,
			formatAmount(t.Amount),
			t.Percent,
			t.TransactionCount,
		)
	}

	_, _ = fmt.Fprintf(stdout, "%-20s %10s %6s %12s\n", "--------------------", "----------", "-----", "------------")
	_, _ = fmt.Fprintf(stdout, "%-20s %10s %5.1f%% %12d\n", "TOTAL", formatAmount(result.ExpenseTotal), 100.0, result.TransactionCount)
}
