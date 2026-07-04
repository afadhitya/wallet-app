package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newAdjustCmd() *cobra.Command {
	var notes string

	cmd := &cobra.Command{
		Use:   "adjust <account> <amount> <description>",
		Short: "Adjust an account balance to a target amount",
		Args:  cobra.ExactArgs(3),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAdjust(cmd, args, svc, notes)
		}),
	}

	cmd.Flags().StringVarP(&notes, "notes", "n", "", "Additional notes for the adjustment")

	return cmd
}

func runAdjust(cmd *cobra.Command, args []string, svc *service.Service, notes string) error {
	account := args[0]
	target, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid amount: %s", args[1]))
	}
	description := args[2]

	result, err := svc.AdjustBalance(service.AdjustBalanceParams{
		Account:     account,
		Target:      target,
		Description: description,
		Notes:       notes,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"account":     result.Account.Name,
			"old_balance": result.OldBalance,
			"new_balance": result.NewBalance,
			"difference":  result.Difference,
			"description": description,
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Balance adjusted for %s:\n", result.Account.Name)
	_, _ = fmt.Fprintf(stdout, "  Old balance: %s\n", formatAmount(result.OldBalance))
	_, _ = fmt.Fprintf(stdout, "  New balance: %s\n", formatAmount(result.NewBalance))

	diffStr := formatAmount(result.Difference)
	if result.Difference > 0 {
		diffStr = "+" + diffStr
	}
	_, _ = fmt.Fprintf(stdout, "  Difference:  %s\n", diffStr)

	return nil
}
