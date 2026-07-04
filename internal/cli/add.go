package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a transaction",
	}

	cmd.AddCommand(newAddExpenseCmd())
	cmd.AddCommand(newAddIncomeCmd())
	cmd.AddCommand(newAddTransferCmd())

	return cmd
}

func newAddExpenseCmd() *cobra.Command {
	var category, account string
	var tags []string
	var date, notes string

	cmd := &cobra.Command{
		Use:   "expense <amount> <description>",
		Short: "Record an expense transaction",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAddExpense(cmd, args, svc, account, category, tags, date, notes)
		}),
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Category name or ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Account name or ID")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Tag name (can be repeated)")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Transaction date (YYYY-MM-DD, today)")
	cmd.Flags().StringVarP(&notes, "notes", "n", "", "Additional notes")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("account")

	return cmd
}

func newAddIncomeCmd() *cobra.Command {
	var category, account string
	var tags []string
	var date, notes string

	cmd := &cobra.Command{
		Use:   "income <amount> <description>",
		Short: "Record an income transaction",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAddIncome(cmd, args, svc, account, category, tags, date, notes)
		}),
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Category name or ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Account name or ID")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Tag name (can be repeated)")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Transaction date (YYYY-MM-DD, today)")
	cmd.Flags().StringVarP(&notes, "notes", "n", "", "Additional notes")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("account")

	return cmd
}

func newAddTransferCmd() *cobra.Command {
	var fromAccount, toAccount string
	var date, notes string

	cmd := &cobra.Command{
		Use:   "transfer <amount>",
		Short: "Record a transfer between accounts",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAddTransfer(cmd, args, svc, fromAccount, toAccount, date, notes)
		}),
	}

	cmd.Flags().StringVar(&fromAccount, "from", "", "Source account name or ID")
	cmd.Flags().StringVar(&toAccount, "to", "", "Destination account name or ID")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Transaction date (YYYY-MM-DD, today)")
	cmd.Flags().StringVarP(&notes, "notes", "n", "", "Additional notes")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func runAddExpense(cmd *cobra.Command, args []string, svc *service.Service, account, category string, tags []string, date, notes string) error {
	amount, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid amount: %s", args[0]))
	}
	description := args[1]

	result, err := svc.AddExpense(service.CreateExpenseParams{
		Amount:      amount,
		Description: description,
		Category:    category,
		Account:     account,
		Tags:        tags,
		Date:        date,
		Notes:       notes,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	return printTransactionResult(cmd, result, "expense")
}

func runAddIncome(cmd *cobra.Command, args []string, svc *service.Service, account, category string, tags []string, date, notes string) error {
	amount, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid amount: %s", args[0]))
	}
	description := args[1]

	result, err := svc.AddIncome(service.CreateIncomeParams{
		Amount:      amount,
		Description: description,
		Category:    category,
		Account:     account,
		Tags:        tags,
		Date:        date,
		Notes:       notes,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	return printTransactionResult(cmd, result, "income")
}

func runAddTransfer(cmd *cobra.Command, args []string, svc *service.Service, fromAccount, toAccount, date, notes string) error {
	amount, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid amount: %s", args[0]))
	}

	result, err := svc.AddTransfer(service.CreateTransferParams{
		Amount:      amount,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Date:        date,
		Notes:       notes,
		Description: fmt.Sprintf("Transfer from %s to %s", fromAccount, toAccount),
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, stderr := resolveOut(cmd)

	if isJSON(cmd) {
		output := map[string]interface{}{
			"id":           result.Transaction.ID,
			"type":         "transfer",
			"amount":       result.Transaction.Amount,
			"from_account": fromAccount,
			"to_account":   toAccount,
			"date":         result.Transaction.Date,
		}
		if result.Warning != "" {
			output["warning"] = result.Warning
		}
		return printJSON(stdout, output)
	}

	_, _ = fmt.Fprintf(stdout, "Transfer recorded: %d from %s to %s on %s\n",
		result.Transaction.Amount, fromAccount, toAccount, result.Transaction.Date)
	if result.Warning != "" {
		_, _ = fmt.Fprintf(stderr, "%s\n", result.Warning)
	}
	return nil
}

func printTransactionResult(cmd *cobra.Command, result *service.TransactionResult, txnType string) error {
	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		desc := ""
		if result.Transaction.Description.Valid {
			desc = result.Transaction.Description.String
		}
		output := map[string]interface{}{
			"id":          result.Transaction.ID,
			"type":        txnType,
			"amount":      result.Transaction.Amount,
			"currency":    result.Transaction.Currency,
			"description": desc,
			"date":        result.Transaction.Date,
			"tags":        tagNames(result.Tags),
		}
		if result.Transaction.BaseAmount.Valid {
			output["base_amount"] = result.Transaction.BaseAmount.Int64
			output["base_currency"] = result.Transaction.BaseCurrency.String
		}
		return printJSON(stdout, output)
	}

	desc := ""
	if result.Transaction.Description.Valid {
		desc = result.Transaction.Description.String
	}
	amountStr := fmt.Sprintf("%d %s", result.Transaction.Amount, result.Transaction.Currency)
	if result.Transaction.BaseAmount.Valid {
		amountStr = fmt.Sprintf("%s (%s %s)", amountStr, formatNum(result.Transaction.BaseAmount.Int64), result.Transaction.BaseCurrency.String)
	}
	_, _ = fmt.Fprintf(stdout, "%s recorded: %s [%s] on %s\n",
		txnType, amountStr, desc, result.Transaction.Date)
	if len(result.Tags) > 0 {
		_, _ = fmt.Fprintf(stdout, "Tags: %v\n", tagNames(result.Tags))
	}
	return nil
}

func tagNames(tags []*gen.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}
