package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/service"
	"github.com/spf13/cobra"
)

func newRateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate",
		Short: "Manage exchange rates",
	}

	cmd.AddCommand(newRateListCmd())
	cmd.AddCommand(newRateSetCmd())
	cmd.AddCommand(newRateAddCmd())
	cmd.AddCommand(newRateRmCmd())

	return cmd
}

func newRateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured exchange rates",
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runRateList(cmd, svc)
		}),
	}
	return cmd
}

func newRateSetCmd() *cobra.Command {
	var currency, rateStr string

	cmd := &cobra.Command{
		Use:   "set <currency> <rate>",
		Short: "Update an existing exchange rate",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runRateSet(cmd, args[0], args[1], svc)
		}),
	}

	_ = currency
	_ = rateStr

	return cmd
}

func newRateAddCmd() *cobra.Command {
	var currency, rateStr string

	cmd := &cobra.Command{
		Use:   "add <currency> <rate>",
		Short: "Add a new exchange rate",
		Args:  cobra.ExactArgs(2),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runRateAdd(cmd, args[0], args[1], svc)
		}),
	}

	_ = currency
	_ = rateStr

	return cmd
}

func newRateRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <currency>",
		Short: "Remove an exchange rate",
		Args:  cobra.ExactArgs(1),
		RunE: withService(func(cmd *cobra.Command, args []string, svc *service.Service, db *sql.DB) error {
			return runRateRm(cmd, args[0], svc)
		}),
	}
	return cmd
}

func runRateList(cmd *cobra.Command, svc *service.Service) error {
	base, rates, err := svc.ListRates()
	if err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)

	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{
			"base_currency": base,
			"rates":         rates,
		})
	}

	if len(rates) == 0 {
		_, _ = fmt.Fprintf(stdout, "No exchange rates configured.\n")
		_, _ = fmt.Fprintf(stdout, "Base currency: %s\n", base)
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Base currency: %s\n\n", base)
	_, _ = fmt.Fprintf(stdout, "%-8s %-15s %-15s\n", "Currency", "Rate (1→base)", "Inverse (1→base)")
	_, _ = fmt.Fprintf(stdout, "%-8s %-15s %-15s\n", "--------", "---------------", "---------------")

	for currency, rate := range rates {
		rateStr := fmt.Sprintf("%s %s", base, formatNum(rate))
		inverse := 1.0 / float64(rate)
		inverseStr := ""
		if inverse < 0.01 {
			inverseStr = fmt.Sprintf("%.6f %s", inverse, currency)
		} else if inverse < 1 {
			inverseStr = fmt.Sprintf("%.4f %s", inverse, currency)
		} else {
			inverseStr = fmt.Sprintf("%.2f %s", inverse, currency)
		}
		_, _ = fmt.Fprintf(stdout, "%-8s %-15s %-15s\n", currency, rateStr, inverseStr)
	}

	return nil
}

func runRateSet(cmd *cobra.Command, currency, rateStr string, svc *service.Service) error {
	rate, err := strconv.ParseInt(rateStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid rate: %s (must be a positive integer)", rateStr))
	}

	if err := svc.SetRate(currency, rate); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	base, _ := svc.GetBaseCurrency()
	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{
			"status":   "updated",
			"currency": currency,
			"rate":     rate,
			"base":     base,
		})
	}
	_, _ = fmt.Fprintf(stdout, "Rate updated: 1 %s = %s %s\n", currency, formatNum(rate), base)
	return nil
}

func runRateAdd(cmd *cobra.Command, currency, rateStr string, svc *service.Service) error {
	rate, err := strconv.ParseInt(rateStr, 10, 64)
	if err != nil {
		return formatError(cmd, fmt.Errorf("invalid rate: %s (must be a positive integer)", rateStr))
	}

	if err := svc.AddRate(currency, rate); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	base, _ := svc.GetBaseCurrency()
	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{
			"status":   "added",
			"currency": currency,
			"rate":     rate,
			"base":     base,
		})
	}
	_, _ = fmt.Fprintf(stdout, "Rate added: 1 %s = %s %s\n", currency, formatNum(rate), base)
	return nil
}

func runRateRm(cmd *cobra.Command, currency string, svc *service.Service) error {
	if err := svc.RemoveRate(currency); err != nil {
		return formatError(cmd, err)
	}

	stdout, _ := resolveOut(cmd)
	if isJSON(cmd) {
		return printJSON(stdout, map[string]interface{}{
			"status":   "removed",
			"currency": currency,
		})
	}
	_, _ = fmt.Fprintf(stdout, "Rate removed: %s\n", currency)
	return nil
}
