package cli

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove (archive) a transaction",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runRm(cmd, args[0], svc, force)
		}),
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runRm(cmd *cobra.Command, idStr string, svc *service.Service, force bool) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid transaction ID: %s", idStr))
	}

	txn, err := svc.GetTransactionByID(id)
	if err != nil {
		return formatError(cmd, err)
	}

	if !force {
		stdout, _ := resolveOut(cmd)

		desc := ""
		if txn.Description.Valid {
			desc = txn.Description.String
		}
		_, _ = fmt.Fprintf(stdout, "Remove transaction #%d: %s %s - %d on %s?\n",
			txn.ID, txn.Type, desc, txn.Amount, txn.Date)
		_, _ = fmt.Fprintf(stdout, "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return formatError(cmd, fmt.Errorf("failed to read confirmation: %w", err))
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input != "yes" && input != "y" {
			stdout, _ := resolveOut(cmd)
			_, _ = fmt.Fprintln(stdout, "Cancelled.")
			return nil
		}
	}

	if err := svc.RemoveTransaction(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{
			"status": "removed",
			"id":     id,
		})
	}

	_, _ = fmt.Fprintf(stdout, "Transaction %d removed.\n", id)
	return nil
}
