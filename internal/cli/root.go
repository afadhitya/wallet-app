package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "A CLI wallet application",
	}

	cmd.PersistentFlags().Bool("json", false, "Enable JSON output for machine-readable results")

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCategoryCmd())
	cmd.AddCommand(newTagCmd())
	cmd.AddCommand(newBudgetCmd())
	cmd.AddCommand(newBillCmd())
	cmd.AddCommand(newReportCmd())
	cmd.AddCommand(newForecastCmd())

	return cmd
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the wallet database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newCategoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newBudgetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "budget",
		Short: "Manage budgets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newBillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bill",
		Short: "Manage bills and planned payments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate financial reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newForecastCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "forecast",
		Short: "Forecast future balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
