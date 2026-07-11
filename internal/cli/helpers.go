package cli

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/afadhitya/wallet-app/internal/db"
	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/afadhitya/wallet-app/internal/service/shared"
	"github.com/afadhitya/wallet-app/pkg/config"
	"github.com/afadhitya/wallet-app/pkg/update"
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

type successMeta struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}

type successResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    successMeta `json:"meta"`
}

func newSuccessResponse(data interface{}, cmd *cobra.Command) *successResponse {
	return &successResponse{
		Success: true,
		Data:    data,
		Meta: successMeta{
			Command:   cmd.CommandPath(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func printSuccessJSON(w io.Writer, data interface{}, cmd *cobra.Command) error {
	return printJSON(w, newSuccessResponse(data, cmd))
}

type errorBody struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type errorResponse struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

const (
	ErrCodeInvalidInput           = "INVALID_INPUT"
	ErrCodeNotFound               = "NOT_FOUND"
	ErrCodeAccountNotFound        = "ACCOUNT_NOT_FOUND"
	ErrCodeCategoryNotFound       = "CATEGORY_NOT_FOUND"
	ErrCodeTagNotFound            = "TAG_NOT_FOUND"
	ErrCodeTransactionNotFound    = "TRANSACTION_NOT_FOUND"
	ErrCodeBudgetNotFound         = "BUDGET_NOT_FOUND"
	ErrCodePlannedPaymentNotFound = "PLANNED_PAYMENT_NOT_FOUND"
	ErrCodeInvalidAmount          = "INVALID_AMOUNT"
	ErrCodeInvalidDate            = "INVALID_DATE"
	ErrCodeValidation             = "VALIDATION_ERROR"
	ErrCodeBillPaused             = "BILL_PAUSED"
	ErrCodeBillAlreadyPaid        = "BILL_ALREADY_PAID"
	ErrCodeExchangeRateNotFound   = "EXCHANGE_RATE_NOT_FOUND"
	ErrCodeExchangeRateConfig     = "EXCHANGE_RATE_CONFIG_MISSING"
	ErrCodeExchangeRateInvalid    = "EXCHANGE_RATE_INVALID"
	ErrCodeDBError                = "DB_ERROR"
	ErrCodeInternal               = "INTERNAL_ERROR"
	ErrCodeUpdateFailed           = "UPDATE_FAILED"
	ErrCodeUpdateChecksumMismatch = "UPDATE_CHECKSUM_MISMATCH"
	ErrCodeUpdateNetworkError     = "UPDATE_NETWORK_ERROR"
	ErrCodeUpdatePermission       = "UPDATE_PERMISSION_ERROR"
	ErrCodeUpdateAlreadyLatest    = "UPDATE_ALREADY_LATEST"
)

func classifyError(err error) (code string, suggestion string) {
	var notFound *shared.NotFoundError
	if errors.As(err, &notFound) {
		switch notFound.Entity {
		case "account":
			return ErrCodeAccountNotFound, ""
		case "category":
			return ErrCodeCategoryNotFound, ""
		case "tag":
			return ErrCodeTagNotFound, ""
		case "transaction":
			return ErrCodeTransactionNotFound, ""
		case "budget":
			return ErrCodeBudgetNotFound, ""
		case "planned payment":
			return ErrCodePlannedPaymentNotFound, ""
		default:
			return ErrCodeNotFound, ""
		}
	}

	var validation *shared.ValidationError
	if errors.As(err, &validation) {
		switch validation.Field {
		case "amount":
			return ErrCodeInvalidAmount, validation.Message
		case "date", "start_date", "from", "to":
			return ErrCodeInvalidDate, validation.Message
		case "state":
			msg := validation.Message
			if strings.Contains(strings.ToLower(msg), "not paused") {
				return ErrCodeBillPaused, "planned payment is not paused"
			}
			if strings.Contains(strings.ToLower(msg), "paused") {
				return ErrCodeBillPaused, "unpause the planned payment first"
			}
			return ErrCodeValidation, msg
		default:
			return ErrCodeValidation, validation.Message
		}
	}

	var rateNotFound *shared.RateNotFoundError
	if errors.As(err, &rateNotFound) {
		return ErrCodeExchangeRateNotFound, fmt.Sprintf("use 'wallet rate add %s <rate>'", rateNotFound.Currency)
	}

	if errors.Is(err, 	shared.ErrInvalidAmount) {
		return ErrCodeInvalidAmount, "amount must be positive"
	}

	if errors.Is(err, 	shared.ErrRateConfigMissing) {
		return ErrCodeExchangeRateConfig, "run 'wallet init' to set up"
	}

	if errors.Is(err, 	shared.ErrRateMustBePositive) {
		return ErrCodeExchangeRateInvalid, "rate must be a positive integer"
	}

	if errors.Is(err, 	shared.ErrDuplicateName) {
		return ErrCodeValidation, "name already exists"
	}

	if errors.Is(err, 	shared.ErrNotFound) {
		return ErrCodeNotFound, ""
	}

	if errors.Is(err, 	shared.ErrMissingField) {
		return ErrCodeValidation, "required field missing"
	}

	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), "database") || strings.Contains(strings.ToLower(msg), "sql") {
		return ErrCodeDBError, ""
	}

	if strings.Contains(strings.ToLower(msg), "invalid") || strings.Contains(strings.ToLower(msg), "required") {
		return ErrCodeInvalidInput, ""
	}

	if errors.Is(err, update.ErrChecksumMismatch) {
		return ErrCodeUpdateChecksumMismatch, ""
	}
	if errors.Is(err, update.ErrNetworkError) {
		return ErrCodeUpdateNetworkError, ""
	}
	if errors.Is(err, update.ErrPermission) {
		return ErrCodeUpdatePermission, ""
	}
	if errors.Is(err, update.ErrAlreadyLatest) {
		return ErrCodeUpdateAlreadyLatest, ""
	}
	if errors.Is(err, update.ErrUpdateFailed) {
		return ErrCodeUpdateFailed, ""
	}

	return ErrCodeInternal, ""
}

func printErrJSON(w io.Writer, msg string) {
	_ = printJSON(w, map[string]string{"error": msg})
}

func printErrorJSON(w io.Writer, code, message, suggestion string) error {
	return printJSON(w, &errorResponse{
		Success: false,
		Error: errorBody{
			Code:       code,
			Message:    message,
			Suggestion: suggestion,
		},
	})
}

func formatError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	_, stderr := resolveOut(cmd)
	if isJSON(cmd) {
		code, suggestion := classifyError(err)
		_ = printErrorJSON(stderr, code, err.Error(), suggestion)
	} else {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err.Error())
	}
	return err
}

var (
	svcConfigLoad = config.Load
	svcMkdirAll   = os.MkdirAll
	svcDBOpen     = db.Open
	svcDBMigrate  = db.Migrate
)

func getService(cmd *cobra.Command) (*service.Service, *sql.DB, error) {
	cfg, err := svcConfigLoad("")
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	dbPath := expandHomePath(cfg.Database.Path)

	dir := filepath.Dir(dbPath)
	if err := svcMkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("create data directory: %w", err)
	}

	database, err := svcDBOpen(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	if err := svcDBMigrate(database); err != nil {
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

var osStdin io.Reader = os.Stdin

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
			defer func() { _ = database.Close() }()
		}
		return f(cmd, args, svc, database)
	}
}
