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
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newRmCmd())
	cmd.AddCommand(newCategoryCmd())
	cmd.AddCommand(newTagCmd())
	cmd.AddCommand(newAdjustCmd())
	cmd.AddCommand(newBudgetCmd())
	cmd.AddCommand(newBillCmd())
	cmd.AddCommand(newReportCmd())
	cmd.AddCommand(newForecastCmd())

	return cmd
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
