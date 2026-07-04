package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/afadhitya/wallet-app/pkg/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the wallet database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd)
		},
	}
}

func runInit(cmd *cobra.Command) error {
	stdout, _ := resolveOut(cmd)

	cfg, err := config.Load("")
	if err != nil {
		return formatError(cmd, fmt.Errorf("load config: %w", err))
	}

	dbPath := expandHomePath(cfg.Database.Path)

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return formatError(cmd, fmt.Errorf("create data directory: %w", err))
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return formatError(cmd, fmt.Errorf("open database: %w", err))
	}
	defer func() { _ = database.Close() }()

	if err := db.Migrate(database); err != nil {
		return formatError(cmd, fmt.Errorf("migrate database: %w", err))
	}

	svc := service.New(database)

	accounts, _ := svc.ListAccounts()
	categories, _ := svc.ListAllCategories()

	if err := config.EnsureRatesFile(); err != nil {
		return formatError(cmd, fmt.Errorf("create rate config: %w", err))
	}

	if getServiceOverride != nil {
		_ = database.Close()
	}

	if isJSON(cmd) {
		_ = printSuccessJSON(stdout, map[string]interface{}{
			"status":     "initialized",
			"database":   dbPath,
			"accounts":   len(accounts),
			"categories": len(categories),
			"message":    "Wallet database initialized successfully",
		}, cmd)
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Wallet initialized successfully.\n")
	_, _ = fmt.Fprintf(stdout, "Database: %s\n", dbPath)
	_, _ = fmt.Fprintf(stdout, "Accounts: %d\n", len(accounts))
	_, _ = fmt.Fprintf(stdout, "Categories: %d\n", len(categories))

	return nil
}
