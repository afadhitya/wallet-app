package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newBudgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage budgets",
	}

	cmd.AddCommand(newBudgetSetCmd())
	cmd.AddCommand(newBudgetListCmd())
	cmd.AddCommand(newBudgetCheckCmd())
	cmd.AddCommand(newBudgetEditCmd())
	cmd.AddCommand(newBudgetRmCmd())

	return cmd
}

func newBudgetSetCmd() *cobra.Command {
	var categories, tags []string
	var allCategories bool
	var period, from, to string
	var notifyPct int64

	cmd := &cobra.Command{
		Use:   "set <name> <amount>",
		Short: "Create or update a budget",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBudgetSet(cmd, args[0], args[1], svc, categories, tags, period, from, to, notifyPct, allCategories)
		}),
	}

	cmd.Flags().StringSliceVarP(&categories, "category", "c", nil, "Category name or ID (can be repeated)")
	cmd.Flags().BoolVarP(&allCategories, "all-categories", "A", false, "Include all expense categories")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Tag name or ID (can be repeated)")
	cmd.Flags().StringVar(&period, "period", "monthly", "Budget period: monthly, weekly, yearly, one_time")
	cmd.Flags().StringVar(&from, "from", "", "Period start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "Period end date (YYYY-MM-DD)")
	cmd.Flags().Int64Var(&notifyPct, "notify", 80, "Notification threshold percentage (1-100)")

	cmd.MarkFlagsMutuallyExclusive("category", "all-categories")

	return cmd
}

func newBudgetListCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List budgets",
		Args:  cobra.NoArgs,
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBudgetList(cmd, svc, all)
		}),
	}

	cmd.Flags().BoolVar(&all, "all", false, "Include inactive budgets")

	return cmd
}

func newBudgetCheckCmd() *cobra.Command {
	var budget string
	var all bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check budget spending progress",
		Args:  cobra.NoArgs,
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBudgetCheck(cmd, svc, budget, all)
		}),
	}

	cmd.Flags().StringVarP(&budget, "budget", "b", "", "Budget ID or name to check")
	cmd.Flags().BoolVar(&all, "all", false, "Check all active budgets")

	return cmd
}

func newBudgetEditCmd() *cobra.Command {
	var amountStr, name string
	var notifyPct int64
	var addCategories, removeCategories, addTags, removeTags []string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a budget",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBudgetEdit(cmd, args[0], svc, amountStr, name, notifyPct, addCategories, removeCategories, addTags, removeTags)
		}),
	}

	cmd.Flags().StringVar(&amountStr, "amount", "", "New budget amount")
	cmd.Flags().StringVar(&name, "name", "", "New budget name")
	cmd.Flags().Int64Var(&notifyPct, "notify", 0, "New notification threshold (1-100)")
	cmd.Flags().StringSliceVar(&addCategories, "add-category", nil, "Category to add (can be repeated)")
	cmd.Flags().StringSliceVar(&removeCategories, "remove-category", nil, "Category to remove (can be repeated)")
	cmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "Tag to add (can be repeated)")
	cmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "Tag to remove (can be repeated)")

	return cmd
}

func newBudgetRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove (deactivate) a budget",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runBudgetRm(cmd, args[0], svc)
		}),
	}
}

func runBudgetSet(cmd *cobra.Command, name, amountStr string, svc *service.Service, categories, tags []string, period, from, to string, notifyPct int64, allCategories bool) error {
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid amount: %s", amountStr))
	}

	params := service.SetBudgetParams{
		Name:          name,
		Amount:        amount,
		Period:        period,
		From:          from,
		To:            to,
		NotifyPct:     notifyPct,
		Categories:    categories,
		AllCategories: allCategories,
		Tags:          tags,
	}

	result, err := svc.SetBudget(params)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"id":             result.Budget.ID,
			"name":           budgetDisplayName(result.Budget),
			"amount":         result.Budget.Amount,
			"currency":       result.Budget.Currency,
			"period":         result.Budget.Type,
			"period_start":   result.Budget.PeriodStart,
			"period_end":     result.Budget.PeriodEnd,
			"notify_at_pct":  budgetNotifyPct(result.Budget),
			"is_active":      result.Budget.IsActive == 1,
			"all_categories": result.Budget.AllCategories == 1,
			"categories":     categoryNames(result.Categories),
			"tags":           tagNames(result.Tags),
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Budget '%s' set: %s (period: %s)\n",
		budgetDisplayName(result.Budget), formatAmount(result.Budget.Amount), result.Budget.Type)
	return nil
}

func runBudgetList(cmd *cobra.Command, svc *service.Service, all bool) error {
	items, err := svc.ListBudgets(service.ListBudgetsParams{All: all})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		output := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			output = append(output, map[string]interface{}{
				"id":             item.Budget.ID,
				"name":           budgetDisplayName(item.Budget),
				"amount":         item.Budget.Amount,
				"spent":          item.Spent,
				"remaining":      item.Remaining,
				"period":         item.Budget.Type,
				"period_start":   item.Budget.PeriodStart,
				"period_end":     item.Budget.PeriodEnd,
				"notify_at_pct":  budgetNotifyPct(item.Budget),
				"is_active":      item.Budget.IsActive == 1,
				"all_categories": item.Budget.AllCategories == 1,
				"categories":     categoryNames(item.Categories),
				"tags":           tagNames(item.Tags),
			})
		}
		return printSuccessJSON(stdout, map[string]interface{}{"budgets": output}, cmd)
	}

	if len(items) == 0 {
		_, _ = fmt.Fprintln(stdout, "No budgets found.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-12s %-12s %s\n",
		"ID", "Name", "Limit", "Spent", "Remaining", "Period")
	for _, item := range items {
		active := ""
		if item.Budget.IsActive == 0 {
			active = " (inactive)"
		}
		_, _ = fmt.Fprintf(stdout, "%-4d %-25s %-12s %-12s %-12s %s%s\n",
			item.Budget.ID,
			truncate(budgetDisplayName(item.Budget), 24),
			formatAmount(item.Budget.Amount),
			formatAmount(item.Spent),
			formatAmount(item.Remaining),
			item.Budget.Type,
			active,
		)
	}
	return nil
}

func runBudgetCheck(cmd *cobra.Command, svc *service.Service, budget string, all bool) error {
	if budget == "" && !all {
		return formatError(cmd, fmt.Errorf("specify --budget or --all"))
	}

	results, err := svc.CheckBudgets(service.CheckBudgetsParams{
		Identifier: budget,
		All:        all,
	})
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		output := make([]map[string]interface{}, 0, len(results))
		for _, r := range results {
			output = append(output, map[string]interface{}{
				"id":            r.Budget.ID,
				"name":          budgetDisplayName(r.Budget),
				"limit":         r.Budget.Amount,
				"spent":         r.Spent,
				"remaining":     r.Remaining,
				"percent_used":  r.PercentUsed,
				"status":        r.Status,
				"period_start":  r.Budget.PeriodStart,
				"period_end":    r.Budget.PeriodEnd,
				"notify_at_pct": budgetNotifyPct(r.Budget),
			})
		}
		return printSuccessJSON(stdout, map[string]interface{}{"budgets": output}, cmd)
	}

	if len(results) == 0 {
		_, _ = fmt.Fprintln(stdout, "No budgets to check.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-12s %-12s %-12s %6s %s\n",
		"ID", "Name", "Limit", "Spent", "Remaining", "% Used", "Status")
	for _, r := range results {
		_, _ = fmt.Fprintf(stdout, "%-4d %-25s %-12s %-12s %-12s %5.0f%% %s\n",
			r.Budget.ID,
			truncate(budgetDisplayName(r.Budget), 24),
			formatAmount(r.Budget.Amount),
			formatAmount(r.Spent),
			formatAmount(r.Remaining),
			r.PercentUsed,
			r.Status,
		)
	}
	return nil
}

func runBudgetEdit(cmd *cobra.Command, idStr string, svc *service.Service, amountStr, name string, notifyPct int64, addCategories, removeCategories, addTags, removeTags []string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid budget ID: %s", idStr))
	}

	params := service.EditBudgetParams{
		Name:             name,
		AddCategories:    addCategories,
		RemoveCategories: removeCategories,
		AddTags:          addTags,
		RemoveTags:       removeTags,
	}

	if amountStr != "" {
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return formatError(cmd, fmt.Errorf("invalid amount: %s", amountStr))
		}
		params.Amount = &amount
	}

	if notifyPct != 0 || cmd.Flags().Changed("notify") {
		pct := notifyPct
		params.NotifyPct = &pct
	}

	result, err := svc.EditBudget(id, params)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{
			"id":             result.Budget.ID,
			"name":           budgetDisplayName(result.Budget),
			"amount":         result.Budget.Amount,
			"period":         result.Budget.Type,
			"period_start":   result.Budget.PeriodStart,
			"period_end":     result.Budget.PeriodEnd,
			"notify_at_pct":  budgetNotifyPct(result.Budget),
			"is_active":      result.Budget.IsActive == 1,
			"all_categories": result.Budget.AllCategories == 1,
			"categories":     categoryNames(result.Categories),
			"tags":           tagNames(result.Tags),
		}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Budget %d updated.\n", result.Budget.ID)
	return nil
}

func runBudgetRm(cmd *cobra.Command, idStr string, svc *service.Service) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid budget ID: %s", idStr))
	}

	if err := svc.RemoveBudget(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printSuccessJSON(stdout, map[string]interface{}{"status": "removed", "id": id}, cmd)
	}

	_, _ = fmt.Fprintf(stdout, "Budget %d removed.\n", id)
	return nil
}

func budgetDisplayName(b *gen.Budget) string {
	if b.Name.Valid {
		return b.Name.String
	}
	return fmt.Sprintf("(id:%d)", b.ID)
}

func budgetNotifyPct(b *gen.Budget) int64 {
	if b.NotifyAtPct.Valid {
		return b.NotifyAtPct.Int64
	}
	return 80
}

func categoryNames(categories []*gen.Category) []string {
	names := make([]string, 0, len(categories))
	for _, c := range categories {
		names = append(names, c.Name)
	}
	return names
}
