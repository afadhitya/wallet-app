package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocsCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation",
		Long:   "Generate documentation for the wallet CLI. This command is hidden from user-facing help.",
		Hidden: true,
	}

	var outputDir string

	markdownCmd := &cobra.Command{
		Use:   "markdown",
		Short: "Generate Markdown CLI reference documentation",
		Long:  "Generate one Markdown file per visible command in the directory specified by --output.",
		RunE: func(mdCmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(root, outputDir)
		},
	}

	markdownCmd.Flags().StringVarP(&outputDir, "output", "o", "docs/cli", "Output directory for generated Markdown files")

	cmd.AddCommand(markdownCmd)

	return cmd
}
