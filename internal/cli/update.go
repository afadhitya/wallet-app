package cli

import (
	"fmt"

	"github.com/afadhitya/wallet-app/pkg/update"
	"github.com/spf13/cobra"
)

type updateOutput struct {
	Previous string `json:"previous"`
	Current  string `json:"current,omitempty"`
}

func newUpdateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update wallet to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already at latest")
	cmd.Flags().Bool("json", false, "Enable JSON output")
	return cmd
}

func runUpdate(cmd *cobra.Command, force bool) error {
	stdout, _ := resolveOut(cmd)
	current := update.CurrentVersion()

	latest, err := update.LatestRelease()
	if err != nil {
		return formatError(cmd, fmt.Errorf("%w: %v", update.ErrNetworkError, err))
	}

	latestVer := latest.TagName

	if !force && !update.IsNewer(current, latestVer) {
		err := update.ErrAlreadyLatest
		if isJSON(cmd) {
			_ = printSuccessJSON(stdout, updateOutput{
				Previous: current,
				Current:  current,
			}, cmd)
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "Already at latest version %s\n", current)
		return err
	}

	binary, err := update.DownloadAndVerify(latest)
	if err != nil {
		return formatError(cmd, fmt.Errorf("%w: %v", update.ErrUpdateFailed, err))
	}

	if err := update.ReplaceBinary(binary); err != nil {
		return formatError(cmd, fmt.Errorf("%w: %v", update.ErrPermission, err))
	}

	if isJSON(cmd) {
		return printSuccessJSON(stdout, updateOutput{
			Previous: current,
			Current:  latestVer,
		}, cmd)
	}
	_, _ = fmt.Fprintf(stdout, "Updated from %s to %s\n", current, latestVer)
	return nil
}
