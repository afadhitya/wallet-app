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
	cmd.AddCommand(newAccountCmd())
	cmd.AddCommand(newAdjustCmd())
	cmd.AddCommand(newBudgetCmd())
	cmd.AddCommand(newBillCmd())
	cmd.AddCommand(newRateCmd())
	cmd.AddCommand(newReportCmd())
	cmd.AddCommand(newForecastCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newUpdateCmd())

	cmd.AddCommand(newDocsCmd(cmd))

	return cmd
}
