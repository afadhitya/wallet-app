package cli

import (
	"database/sql"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newBillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bill",
		Short: "Manage bills and planned payments",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return cmd.Help()
		}),
	}

	cmd.AddCommand(newBillAddCmd())
	cmd.AddCommand(newBillListCmd())
	cmd.AddCommand(newBillDueCmd())
	cmd.AddCommand(newBillPayCmd())
	cmd.AddCommand(newBillSkipCmd())
	cmd.AddCommand(newBillPauseCmd())
	cmd.AddCommand(newBillResumeCmd())
	cmd.AddCommand(newBillEditCmd())
	cmd.AddCommand(newBillRmCmd())

	return cmd
}

func newBillAddCmd() *cobra.Command {
	var category, account string
	var daily, weekly, monthly, yearly, custom bool
	var rrule string
	var startDate string
	var day int

	cmd := &cobra.Command{
		Use:   "add <name> <amount>",
		Short: "Add a planned payment",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBillAdd(cmd, svc, args[0], args[1], category, account, daily, weekly, monthly, yearly, custom, rrule, startDate, day)
		}),
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Category name or ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Account name or ID")
	cmd.Flags().BoolVar(&daily, "daily", false, "Daily recurrence")
	cmd.Flags().BoolVar(&weekly, "weekly", false, "Weekly recurrence")
	cmd.Flags().BoolVar(&monthly, "monthly", false, "Monthly recurrence")
	cmd.Flags().BoolVar(&yearly, "yearly", false, "Yearly recurrence")
	cmd.Flags().BoolVar(&custom, "custom", false, "Custom recurrence (requires --rrule)")
	cmd.Flags().StringVar(&rrule, "rrule", "", "Recurrence rule (RFC 5545, e.g., FREQ=MONTHLY;BYMONTHDAY=15)")
	cmd.Flags().StringVar(&startDate, "from", "", "Start date (YYYY-MM-DD, default: today)")
	cmd.Flags().IntVar(&day, "day", 0, "Due day of the month/week (e.g., 15 for monthly)")

	return cmd
}

func runBillAdd(cmd *cobra.Command, svc *service.Service, name, amountStr string, category, account string, daily, weekly, monthly, yearly, custom bool, rrule, startDate string, day int) error {
	amount, err := parseAmountArg(amountStr)
	if err != nil {
		return formatError(cmd, err)
	}

	recurrence, err := parseRecurrenceFlags(daily, weekly, monthly, yearly, custom)
	if err != nil {
		return formatError(cmd, err)
	}

	pp, err := svc.CreatePlannedPayment(service.CreatePlannedPaymentParams{
		Name:           name,
		Amount:         amount,
		Account:        account,
		Category:       category,
		Recurrence:     recurrence,
		RecurrenceRule: rrule,
		StartDate:      startDate,
		DueDay:         day,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, pp, cmd)
	}

	nextDue := "N/A"
	if pp.NextDueDate.Valid {
		nextDue = pp.NextDueDate.String
	}

	_, _ = fmt.Fprintf(stdout, "Created planned payment #%d: %s\n", pp.ID, pp.Name)
	_, _ = fmt.Fprintf(stdout, "  Amount: %s | Recurrence: %s | Next due: %s\n",
		formatAmount(pp.Amount), pp.Recurrence, nextDue)
	return nil
}

func parseAmountArg(amountStr string) (int64, error) {
	var amount int64
	_, err := fmt.Sscanf(amountStr, "%d", &amount)
	if err != nil {
		return 0, &service.ValidationError{Field: "amount", Message: "invalid amount"}
	}
	return amount, nil
}

func parseRecurrenceFlags(daily, weekly, monthly, yearly, custom bool) (string, error) {
	flags := map[string]bool{"daily": daily, "weekly": weekly, "monthly": monthly, "yearly": yearly, "custom": custom}
	count := 0
	result := "none"
	for k, v := range flags {
		if v {
			count++
			result = k
		}
	}
	if count > 1 {
		return "", &service.ValidationError{Field: "recurrence", Message: "only one recurrence flag can be set"}
	}
	return result, nil
}

func newBillListCmd() *cobra.Command {
	var paused, all, active bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List planned payments",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBillList(cmd, svc, active, paused, all)
		}),
	}

	cmd.Flags().BoolVar(&paused, "paused", false, "Show paused planned payments")
	cmd.Flags().BoolVar(&all, "all", false, "Show all (active, paused, archived)")
	cmd.Flags().BoolVar(&active, "active", false, "Show active planned payments (default)")

	return cmd
}

func runBillList(cmd *cobra.Command, svc *service.Service, active, paused, all bool) error {
	includePaused := paused || all
	includeArchived := all

	payments, err := svc.ListPlannedPayments(includePaused, includeArchived)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"planned_payments": payments,
			"count":            len(payments),
		}, cmd)
	}

	if len(payments) == 0 {
		_, _ = fmt.Fprintln(stdout, "No planned payments found.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "%-6s %-22s %-20s %-10s %-12s %-8s\n", "ID", "Name", "Next Due", "Amount", "Recurrence", "Status")
	_, _ = fmt.Fprintf(stdout, "%-6s %-22s %-20s %-10s %-12s %-8s\n", "------", "----------------------", "--------------------", "----------", "------------", "--------")

	for _, pp := range payments {
		nextDue := "N/A"
		if pp.NextDueDate.Valid {
			nextDue = pp.NextDueDate.String
		}
		status := "active"
		if pp.IsPaused != 0 {
			status = "paused"
		}
		if pp.IsActive == 0 {
			status = "archived"
		}
		_, _ = fmt.Fprintf(stdout, "%-6d %-22s %-20s %-10s %-12s %-8s\n",
			pp.ID, truncate(pp.Name, 22), nextDue, formatAmount(pp.Amount), pp.Recurrence, status)
	}

	return nil
}

func newBillDueCmd() *cobra.Command {
	var overdue, week bool
	var next int

	cmd := &cobra.Command{
		Use:   "due",
		Short: "Show due planned payments",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBillDue(cmd, svc, overdue, week, next)
		}),
	}

	cmd.Flags().BoolVar(&overdue, "overdue", false, "Show overdue planned payments")
	cmd.Flags().BoolVar(&week, "week", false, "Show payments due this week")
	cmd.Flags().IntVar(&next, "next", 0, "Show payments due in the next N days")

	return cmd
}

func runBillDue(cmd *cobra.Command, svc *service.Service, overdue, week bool, next int) error {
	var filter service.ListDueFilter
	switch {
	case overdue:
		filter = service.DueOverdue
	case week:
		filter = service.DueCurrentWeek
	case next > 0:
		filter = service.DueNextDays
	default:
		filter = service.DueCurrentMonth
	}

	due, total, err := svc.ListDuePlannedPayments(service.ListDueParams{
		Filter:   filter,
		NextDays: next,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"due":       due,
			"total_due": total,
			"count":     len(due),
		}, cmd)
	}

	if len(due) == 0 {
		_, _ = fmt.Fprintln(stdout, "No due planned payments.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "%-6s %-22s %-12s %-10s %-12s\n", "ID", "Name", "Due Date", "Amount", "Recurrence")
	_, _ = fmt.Fprintf(stdout, "%-6s %-22s %-12s %-10s %-12s\n", "------", "----------------------", "------------", "----------", "------------")

	for _, pp := range due {
		nextDue := "N/A"
		if pp.NextDueDate.Valid {
			nextDue = pp.NextDueDate.String
		}
		_, _ = fmt.Fprintf(stdout, "%-6d %-22s %-12s %-10s %-12s\n",
			pp.ID, truncate(pp.Name, 22), nextDue, formatAmount(pp.Amount), pp.Recurrence)
	}

	_, _ = fmt.Fprintf(stdout, "\nTotal due: %s across %d payment(s)\n", formatAmount(total), len(due))

	return nil
}

func newBillPayCmd() *cobra.Command {
	var date string
	var amount int64

	cmd := &cobra.Command{
		Use:   "pay <id>",
		Short: "Pay a planned payment",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillPay(cmd, svc, id, date, amount)
		}),
	}

	cmd.Flags().StringVar(&date, "date", "", "Payment date (YYYY-MM-DD, default: today)")
	cmd.Flags().Int64Var(&amount, "amount", 0, "Override payment amount (default: planned amount)")

	return cmd
}

func runBillPay(cmd *cobra.Command, svc *service.Service, id int64, date string, amount int64) error {
	result, err := svc.PayPlannedPayment(service.PayPlannedPaymentParams{
		ID:     id,
		Date:   date,
		Amount: amount,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, result, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Paid planned payment #%d: %s\n", result.PlannedPayment.ID, result.PlannedPayment.Name)
	_, _ = fmt.Fprintf(stdout, "  Transaction: #%d | Amount: %s\n",
		result.Transaction.ID, formatAmount(result.Transaction.Amount))
	if result.NextDueDate != "" {
		_, _ = fmt.Fprintf(stdout, "  Next due: %s\n", result.NextDueDate)
	} else if result.PlannedPayment.IsActive == 0 {
		_, _ = fmt.Fprintf(stdout, "  Bill archived (one-time payment)\n")
	}
	return nil
}

func newBillSkipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip <id>",
		Short: "Skip a planned payment occurrence",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillSkip(cmd, svc, id)
		}),
	}

	return cmd
}

func runBillSkip(cmd *cobra.Command, svc *service.Service, id int64) error {
	pp, err := svc.SkipPlannedPayment(id)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, pp, cmd)
	}

	nextDue := "N/A"
	if pp.NextDueDate.Valid {
		nextDue = pp.NextDueDate.String
	}
	_, _ = fmt.Fprintf(stdout, "Skipped planned payment #%d: %s\n", pp.ID, pp.Name)
	_, _ = fmt.Fprintf(stdout, "  Next due: %s\n", nextDue)
	return nil
}

func newBillPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause a planned payment",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillPause(cmd, svc, id)
		}),
	}

	return cmd
}

func runBillPause(cmd *cobra.Command, svc *service.Service, id int64) error {
	if err := svc.PausePlannedPayment(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	pp, _ := svc.GetPlannedPaymentByID(id)
	if isJSON(cmd) {
		if pp != nil {
			return printSuccessJSON(stdout, pp, cmd)
		}
		return printSuccessJSON(stdout, map[string]string{"status": "paused", "id": fmt.Sprintf("%d", id)}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Paused planned payment #%d\n", id)
	return nil
}

func newBillResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume a paused planned payment",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillResume(cmd, svc, id)
		}),
	}

	return cmd
}

func runBillResume(cmd *cobra.Command, svc *service.Service, id int64) error {
	if err := svc.ResumePlannedPayment(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	pp, _ := svc.GetPlannedPaymentByID(id)
	if isJSON(cmd) {
		if pp != nil {
			return printSuccessJSON(stdout, pp, cmd)
		}
		return printSuccessJSON(stdout, map[string]string{"status": "resumed", "id": fmt.Sprintf("%d", id)}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Resumed planned payment #%d\n", id)
	return nil
}

func newBillEditCmd() *cobra.Command {
	var name, category, account, rrule, startDate, recurrenceStr string
	var amount int64
	var day int

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a planned payment",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillEdit(cmd, svc, id, name, amount, category, account, rrule, recurrenceStr, startDate, day)
		}),
	}

	cmd.Flags().StringVar(&name, "name", "", "New name")
	cmd.Flags().Int64Var(&amount, "amount", 0, "New amount")
	cmd.Flags().StringVarP(&category, "category", "c", "", "New category name or ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "New account name or ID")
	cmd.Flags().StringVar(&rrule, "rrule", "", "New recurrence rule")
	cmd.Flags().StringVar(&recurrenceStr, "recurrence", "", "New recurrence type (none, daily, weekly, monthly, yearly, custom)")
	cmd.Flags().StringVar(&startDate, "from", "", "New start date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&day, "day", 0, "New due day of the month/week")

	return cmd
}

func runBillEdit(cmd *cobra.Command, svc *service.Service, id int64, name string, amount int64, category, account, rrule, recurrenceStr, startDate string, day int) error {
	params := service.EditPlannedPaymentParams{}

	if cmd.Flags().Changed("name") {
		params.Name = &name
	}
	if cmd.Flags().Changed("amount") {
		params.Amount = &amount
	}
	if cmd.Flags().Changed("category") {
		params.Category = &category
	}
	if cmd.Flags().Changed("account") {
		params.Account = &account
	}
	if cmd.Flags().Changed("rrule") {
		params.RecurrenceRule = &rrule
	}
	if cmd.Flags().Changed("recurrence") {
		params.Recurrence = &recurrenceStr
	}
	if cmd.Flags().Changed("from") {
		params.StartDate = &startDate
	}
	if cmd.Flags().Changed("day") {
		params.DueDay = &day
	}

	pp, err := svc.EditPlannedPayment(id, params)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, pp, cmd)
	}

	nextDue := "N/A"
	if pp.NextDueDate.Valid {
		nextDue = pp.NextDueDate.String
	}
	_, _ = fmt.Fprintf(stdout, "Updated planned payment #%d: %s\n", pp.ID, pp.Name)
	_, _ = fmt.Fprintf(stdout, "  Amount: %s | Recurrence: %s | Next due: %s\n",
		formatAmount(pp.Amount), pp.Recurrence, nextDue)
	return nil
}

func newBillRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Delete a planned payment",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			id, err := parseIDArg(args[0])
			if err != nil {
				return formatError(cmd, err)
			}
			return runBillRm(cmd, svc, id)
		}),
	}

	return cmd
}

func runBillRm(cmd *cobra.Command, svc *service.Service, id int64) error {
	if err := svc.DeletePlannedPayment(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"status": "deleted",
			"id":     id,
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Deleted planned payment #%d\n", id)
	return nil
}

func parseIDArg(arg string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(arg, "%d", &id)
	if err != nil {
		return 0, &service.ValidationError{Field: "id", Message: "invalid ID"}
	}
	return id, nil
}
