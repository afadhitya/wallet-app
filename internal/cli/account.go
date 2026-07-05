package cli

import (
	"bufio"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage accounts",
	}

	cmd.AddCommand(newAccountAddCmd())
	cmd.AddCommand(newAccountListCmd())
	cmd.AddCommand(newAccountEditCmd())
	cmd.AddCommand(newAccountArchiveCmd())

	return cmd
}

func newAccountAddCmd() *cobra.Command {
	var accountType, currency string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an account",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAccountAdd(cmd, args[0], svc, accountType, currency)
		}),
	}

	cmd.Flags().StringVarP(&accountType, "type", "t", "checking", "Account type (checking, savings, ewallet, cash, investment, credit, loan, other)")
	cmd.Flags().StringVarP(&currency, "currency", "c", "IDR", "Currency code (e.g., IDR, USD)")

	return cmd
}

func newAccountListCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAccountList(cmd, svc, all)
		}),
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Include archived accounts")

	return cmd
}

func newAccountEditCmd() *cobra.Command {
	var name, accountType string
	var sortOrder int64

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an account",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAccountEdit(cmd, args[0], svc, name, accountType, sortOrder)
		}),
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "New account name")
	cmd.Flags().StringVarP(&accountType, "type", "t", "", "New account type")
	cmd.Flags().Int64Var(&sortOrder, "sort-order", 0, "Sort order")

	return cmd
}

func newAccountArchiveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive an account",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runAccountArchive(cmd, args[0], svc, force)
		}),
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runAccountAdd(cmd *cobra.Command, name string, svc *service.Service, accountType, currency string) error {
	result, err := svc.CreateAccount(name, accountType, currency)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, result, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Account '%s' created (ID: %d).\n", result.Name, result.ID)
	return nil
}

func runAccountList(cmd *cobra.Command, svc *service.Service, all bool) error {
	var accounts []*gen.Account
	var err error

	if all {
		accounts, err = svc.ListAllAccounts()
	} else {
		accounts, err = svc.ListAccounts()
	}
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, stderr := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, accounts, cmd)
	}

	if len(accounts) == 0 {
		_, _ = fmt.Fprintln(stdout, "No accounts found.")
		return nil
	}

	baseCurrency, rates, err := svc.ListRates()
	if err != nil {
		return formatError(cmd, err)
	}

	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-8s %-15s %-15s %s\n", "ID", "Name", "Type", "Currency", "Balance", "Converted", "Status")
	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-8s %-15s %-15s %s\n", "----", "-------------------------", "------------", "--------", "---------------", "---------------", "------")

	var totalBalance int64
	var missingRates []string
	for _, acc := range accounts {
		status := "active"
		if acc.IsArchived == 1 {
			status = "archived"
		}

		var convertedStr string
		if acc.Currency == baseCurrency {
			convertedStr = "-"
			totalBalance += acc.Balance
		} else if rate, ok := rates[acc.Currency]; ok {
			convertedStr = formatAmount(acc.Balance * rate)
			totalBalance += acc.Balance * rate
		} else {
			convertedStr = "-"
			missingRates = append(missingRates, acc.Currency)
		}

		_, _ = fmt.Fprintf(stdout, "%-4d %-25s %-12s %-8s %-15s %-15s %s\n",
			acc.ID, truncate(acc.Name, 25), acc.Type, acc.Currency, formatAmount(acc.Balance), convertedStr, status)
	}

	totalLabel := fmt.Sprintf("Total (%s):", baseCurrency)
	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-8s %-15s %-15s\n", "----", "-------------------------", "------------", "--------", "---------------", "---------------")
	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-8s %s\n", "", totalLabel, "", "", formatAmount(totalBalance))

	if len(missingRates) > 0 {
		_, _ = fmt.Fprintf(stderr, "Warning: no exchange rate configured for: %s\n", strings.Join(missingRates, ", "))
		_, _ = fmt.Fprintf(stderr, "These accounts are excluded from the total.\n")
	}

	return nil
}

func runAccountEdit(cmd *cobra.Command, idStr string, svc *service.Service, name, accountType string, sortOrder int64) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid account ID: %s", idStr))
	}

	if name == "" && accountType == "" && sortOrder == 0 {
		return formatError(cmd, fmt.Errorf("at least one field (--name, --type, --sort-order) must be specified"))
	}

	existing, err := svc.GetAccountByID(id)
	if err != nil {
		return formatError(cmd, err)
	}

	if name == "" {
		name = existing.Name
	}
	if accountType == "" {
		accountType = existing.Type
	}

	result, err := svc.UpdateAccount(id, name, accountType, existing.Currency, sortOrder)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, result, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Account %d updated.\n", result.ID)
	return nil
}

func runAccountArchive(cmd *cobra.Command, idStr string, svc *service.Service, force bool) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid account ID: %s", idStr))
	}

	account, err := svc.GetAccountByID(id)
	if err != nil {
		return formatError(cmd, err)
	}

	if !force {
		stdout, _ := resolveOut(cmd)

		if account.Balance != 0 {
			_, _ = fmt.Fprintf(stdout, "Warning: Account '%s' has a non-zero balance of %s.\n", account.Name, formatAmount(account.Balance))
		}

		_, _ = fmt.Fprintf(stdout, "Archive account '%s' (ID: %d)?\nType 'yes' to confirm: ", account.Name, account.ID)

		reader := bufio.NewReader(osStdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return formatError(cmd, fmt.Errorf("failed to read confirmation: %w", err))
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input != "yes" && input != "y" {
			_, _ = fmt.Fprintln(stdout, "Cancelled.")
			return nil
		}
	}

	if err := svc.ArchiveAccount(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"status": "archived",
			"id":     id,
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Account %d archived.\n", id)
	return nil
}
