package cli

import (
	"fmt"

	"github.com/afadhitya/wallet-app/pkg/update"
	"github.com/spf13/cobra"
)

type versionOutput struct {
	Version         string `json:"version"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"update_available,omitempty"`
}

func newVersionCmd() *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the current wallet version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(cmd, check)
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Check for latest version from GitHub")
	cmd.Flags().Bool("json", false, "Enable JSON output")
	return cmd
}

func runVersion(cmd *cobra.Command, check bool) error {
	stdout, _ := resolveOut(cmd)
	current := update.CurrentVersion()

	if !check {
		if isJSON(cmd) {
			return printSuccessJSON(stdout, versionOutput{Version: current}, cmd)
		}
		_, _ = fmt.Fprintf(stdout, "%s\n", current)
		return nil
	}

	latest, err := update.LatestRelease()
	if err != nil {
		if isJSON(cmd) {
			_ = printSuccessJSON(stdout, versionOutput{
				Version:         current,
				UpdateAvailable: false,
			}, cmd)
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "%s (unable to check for updates: %v)\n", current, err)
		return nil
	}

	latestVer := latest.TagName

	if update.IsNewer(current, latestVer) {
		if isJSON(cmd) {
			return printSuccessJSON(stdout, versionOutput{
				Version:         current,
				Latest:          latestVer,
				UpdateAvailable: true,
			}, cmd)
		}
		_, _ = fmt.Fprintf(stdout, "%s\n", current)
		_, _ = fmt.Fprintf(stdout, "A new version is available: %s\n", latestVer)
		_, _ = fmt.Fprintf(stdout, "Run 'wallet update' to upgrade.\n")
		return nil
	}

	if isJSON(cmd) {
		return printSuccessJSON(stdout, versionOutput{
			Version:         current,
			Latest:          latestVer,
			UpdateAvailable: false,
		}, cmd)
	}
	_, _ = fmt.Fprintf(stdout, "%s (up to date)\n", current)
	return nil
}
