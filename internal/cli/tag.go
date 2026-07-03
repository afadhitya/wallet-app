package cli

import (
	"database/sql"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagAddCmd())
	cmd.AddCommand(newTagRmCmd())

	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tags",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runTagList(cmd, svc)
		}),
	}
}

func newTagAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Add a tag",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runTagAdd(cmd, args[0], svc)
		}),
	}
}

func newTagRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a tag",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runTagRm(cmd, args[0], svc)
		}),
	}
}

func runTagList(cmd *cobra.Command, svc *service.Service) error {
	tags, err := svc.ListTags()
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, tags)
	}

	if len(tags) == 0 {
		_, _ = fmt.Fprintln(stdout, "No tags found.")
		return nil
	}

	for _, t := range tags {
		_, _ = fmt.Fprintf(stdout, "%-4d %s\n", t.ID, t.Name)
	}
	return nil
}

func runTagAdd(cmd *cobra.Command, name string, svc *service.Service) error {
	result, err := svc.CreateTag(name)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, result)
	}

	_, _ = fmt.Fprintf(stdout, "Tag '%s' created (ID: %d).\n", result.Name, result.ID)
	return nil
}

func runTagRm(cmd *cobra.Command, name string, svc *service.Service) error {
	tag, err := svc.GetTagByName(name)
	if err != nil {
		return formatError(cmd, err)
	}

	if err := svc.DeleteTag(tag.ID); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{"status": "removed", "name": name})
	}

	_, _ = fmt.Fprintf(stdout, "Tag '%s' removed.\n", name)
	return nil
}
