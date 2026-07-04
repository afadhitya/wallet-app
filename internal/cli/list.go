package cli

import (
	"database/sql"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var account, category, tag, txnType, month, dateFrom, dateTo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runList(cmd, svc, account, category, tag, txnType, month, dateFrom, dateTo, limit)
		}),
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Filter by month (e.g., july, 2026-07)")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Filter by account name or ID")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category name or ID")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Filter by tag name")
	cmd.Flags().StringVar(&txnType, "type", "", "Filter by transaction type (expense, income, transfer, adjustment)")
	cmd.Flags().StringVar(&dateFrom, "from", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "to", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Maximum transactions to show")

	return cmd
}

func runList(cmd *cobra.Command, svc *service.Service, account, category, tag, txnType, month, dateFrom, dateTo string, limit int) error {
	result, err := svc.ListTransactions(service.ListTransactionsParams{
		AccountName:  account,
		CategoryName: category,
		TagName:      tag,
		Type:         txnType,
		Month:        month,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		Limit:        limit,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		output := map[string]interface{}{
			"transactions": result.Transactions,
			"total":        formatAmount(result.Total),
			"count":        len(result.Transactions),
		}
		if result.BaseTotal != 0 {
			baseCurrency, _ := svc.GetBaseCurrency()
			output["base_total"] = formatAmount(result.BaseTotal)
			output["base_currency"] = baseCurrency
		}
		return printSuccessJSON(stdout, output, cmd)
	}

	if len(result.Transactions) == 0 {
		_, _ = fmt.Fprintln(stdout, "No transactions found.")
		return nil
	}

	baseCurrency, _ := svc.GetBaseCurrency()
	hasBaseTotal := result.BaseTotal != 0

	_, _ = fmt.Fprintf(stdout, "%-6s %-10s %-12s %-20s %-18s %s\n", "ID", "Date", "Type", "Description", "Amount", "Category")
	_, _ = fmt.Fprintf(stdout, "%-6s %-10s %-12s %-20s %-18s %s\n", "------", "----------", "------------", "--------------------", "------------------", "--------")

	for _, t := range result.Transactions {
		desc := ""
		if t.Description.Valid {
			desc = t.Description.String
		}
		categoryName, _ := svc.GetCategoryByID(t.CategoryID.Int64)
		cat := ""
		if categoryName != nil {
			cat = categoryName.Name
		}

		amountStr := fmt.Sprintf("%s %s", formatNum(t.Amount), t.Currency)
		if t.BaseAmount.Valid {
			amountStr = fmt.Sprintf("%s (%s %s)", amountStr, formatNum(t.BaseAmount.Int64), baseCurrency)
		}

		_, _ = fmt.Fprintf(stdout, "%-6d %-10s %-12s %-20s %-18s %s\n",
			t.ID, t.Date, t.Type, truncate(desc, 20), amountStr, cat)
	}

	totalStr := formatAmount(result.Total)
	if hasBaseTotal {
		totalStr = fmt.Sprintf("%s (Base: %s)", totalStr, formatAmount(result.BaseTotal))
	}
	_, _ = fmt.Fprintf(stdout, "\nTotal: %s (%d transactions)\n", totalStr, len(result.Transactions))
	return nil
}
