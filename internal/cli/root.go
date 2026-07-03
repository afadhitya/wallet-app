package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet",
		Short: "A CLI wallet application",
	}
}
