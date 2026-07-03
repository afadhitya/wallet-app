package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/afadhitya/wallet-app/pkg/config"
	"github.com/spf13/cobra"
)

func isJSON(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("json")
	if !v {
		if parent := cmd.Parent(); parent != nil {
			v, _ = parent.PersistentFlags().GetBool("json")
		}
	}
	return v
}

func printJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printErrJSON(w io.Writer, msg string) {
	_ = printJSON(w, map[string]string{"error": msg})
}

func formatError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	_, stderr := resolveOut(cmd)
	if isJSON(cmd) {
		printErrJSON(stderr, err.Error())
	} else {
		fmt.Fprintf(stderr, "Error: %s\n", err.Error())
	}
	return err
}

func getService(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	dbPath := expandHomePath(cfg.Database.Path)

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("create data directory: %w", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Migrate(database); err != nil {
		_ = database.Close()
		return nil, nil, fmt.Errorf("migrate database: %w", err)
	}

	return service.New(database), database, nil
}

func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func resolveOut(cmd *cobra.Command) (io.Writer, io.Writer) {
	return cmd.OutOrStdout(), cmd.ErrOrStderr()
}

type svcFunc func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error

var getServiceOverride func(*cobra.Command) (*service.Service, *sql.DB, error)

func withService(f svcFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, database, err := func() (*service.Service, *sql.DB, error) {
			if getServiceOverride != nil {
				return getServiceOverride(cmd)
			}
			return getService(cmd)
		}()
		if err != nil {
			return formatError(cmd, err)
		}
		if getServiceOverride == nil {
			defer database.Close()
		}
		return f(cmd, args, svc, database)
	}
}
