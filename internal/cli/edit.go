package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var amountStr, category, account, date, description, notes string
	var addTags, removeTags []string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a transaction",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runEdit(cmd, args[0], svc, amountStr, category, account, date, description, notes, addTags, removeTags)
		}),
	}

	cmd.Flags().StringVar(&amountStr, "amount", "", "New amount")
	cmd.Flags().StringVarP(&category, "category", "c", "", "New category name or ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "New account name or ID")
	cmd.Flags().StringVarP(&date, "date", "d", "", "New date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&description, "desc", "", "New description")
	cmd.Flags().StringVarP(&notes, "notes", "n", "", "New notes")
	cmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "Tag to add (can be repeated)")
	cmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "Tag to remove (can be repeated)")

	return cmd
}

func runEdit(cmd *cobra.Command, idStr string, svc *service.Service, amountStr, category, account, date, description, notes string, addTags, removeTags []string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid transaction ID: %s", idStr))
	}

	var amount *int64
	if amountStr != "" {
		a, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return formatError(cmd, fmt.Errorf("invalid amount: %s", amountStr))
		}
		amount = &a
	}

	result, err := svc.EditTransaction(id, service.EditTransactionParams{
		Amount:         amount,
		CategoryName:   category,
		AccountName:    account,
		Date:           date,
		Description:    description,
		Notes:          notes,
		AddTagNames:    addTags,
		RemoveTagNames: removeTags,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		desc := ""
		if result.Transaction.Description.Valid {
			desc = result.Transaction.Description.String
		}
		return printSuccessJSON(stdout, map[string]interface{}{
			"id":          result.Transaction.ID,
			"type":        result.Transaction.Type,
			"amount":      result.Transaction.Amount,
			"description": desc,
			"date":        result.Transaction.Date,
			"tags":        tagNames(result.Tags),
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Transaction %d updated successfully.\n", result.Transaction.ID)
	return nil
}
