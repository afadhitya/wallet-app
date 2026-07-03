package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newCategoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	cmd.AddCommand(newCategoryListCmd())
	cmd.AddCommand(newCategoryAddCmd())
	cmd.AddCommand(newCategoryEditCmd())
	cmd.AddCommand(newCategoryRmCmd())

	return cmd
}

func newCategoryListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List categories",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runCategoryList(cmd, svc)
		}),
	}
}

func newCategoryAddCmd() *cobra.Command {
	var parentStr, icon string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a category",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runCategoryAdd(cmd, args[0], svc, parentStr, icon)
		}),
	}

	cmd.Flags().StringVarP(&parentStr, "parent", "p", "", "Parent category ID")
	cmd.Flags().StringVar(&icon, "icon", "", "Emoji icon for the category")

	return cmd
}

func newCategoryEditCmd() *cobra.Command {
	var name, icon string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a category",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runCategoryEdit(cmd, args[0], svc, name, icon)
		}),
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "New category name")
	cmd.Flags().StringVar(&icon, "icon", "", "New emoji icon")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newCategoryRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove (archive) a category",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runCategoryRm(cmd, args[0], svc)
		}),
	}
}

func runCategoryList(cmd *cobra.Command, svc *service.Service) error {
	categories, err := svc.ListCategories()
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, categories)
	}

	if len(categories) == 0 {
		_, _ = fmt.Fprintln(stdout, "No categories found.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-8s %s\n", "ID", "Name", "Type", "Parent")
	_, _ = fmt.Fprintf(stdout, "%-4s %-25s %-8s %s\n", "----", "-------------------------", "--------", "------")

	for _, cat := range categories {
		icon := ""
		if cat.Icon.Valid {
			icon = cat.Icon.String + " "
		}

		parentName := "-"
		if cat.ParentID.Valid {
			parentName = fmt.Sprintf("%d", cat.ParentID.Int64)
		}

		_, _ = fmt.Fprintf(stdout, "%-4d %s%-25s %-8s %s\n", cat.ID, icon, cat.Name, cat.Type, parentName)
	}
	return nil
}

func runCategoryAdd(cmd *cobra.Command, name string, svc *service.Service, parentStr, icon string) error {
	result, err := svc.CreateCategory(name, parentStr, icon)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, result)
	}

	_, _ = fmt.Fprintf(stdout, "Category '%s' created (ID: %d).\n", result.Name, result.ID)
	return nil
}

func runCategoryEdit(cmd *cobra.Command, idStr string, svc *service.Service, name, icon string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid category ID: %s", idStr))
	}

	result, err := svc.UpdateCategory(id, name, icon)
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, result)
	}

	_, _ = fmt.Fprintf(stdout, "Category %d updated.\n", result.ID)
	return nil
}

func runCategoryRm(cmd *cobra.Command, idStr string, svc *service.Service) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid category ID: %s", idStr))
	}

	if err := svc.ArchiveCategory(id); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{"status": "removed", "id": id})
	}

	_, _ = fmt.Fprintf(stdout, "Category %d removed.\n", id)
	return nil
}
