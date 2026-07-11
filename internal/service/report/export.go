package report

import (
	"context"
	"fmt"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (m *ReportManager) GenerateExportRows(params ReportParams) ([]ReportExportRow, error) {
	filters, err := m.resolveReportFilters(params)
	if err != nil {
		return nil, err
	}

	arg := m.genReportParams(filters)
	ctx := context.Background()

	rows, err := m.q.ReportExportTransactions(ctx, gen.ReportExportTransactionsParams(arg))
	if err != nil {
		return nil, err
	}

	exportRows := make([]ReportExportRow, 0, len(rows))
	for _, r := range rows {
		baseAmount := int64(0)
		if r.BaseAmount.Valid {
			baseAmount = r.BaseAmount.Int64
		}

		categoryName := ""
		if r.CategoryName.Valid {
			categoryName = r.CategoryName.String
		}

		description := ""
		if r.Description.Valid {
			description = r.Description.String
		}

		tags, err := m.q.ListTransactionTags(ctx, r.ID)
		if err != nil {
			return nil, err
		}

		tagNames := make([]string, 0, len(tags))
		for _, t := range tags {
			tagNames = append(tagNames, t.Name)
		}

		exportRows = append(exportRows, ReportExportRow{
			Date:        r.Date,
			Type:        r.Type,
			Amount:      r.Amount,
			Currency:    r.Currency,
			BaseAmount:  baseAmount,
			Category:    categoryName,
			Account:     r.AccountName,
			Description: description,
			Tags:        joinTagNames(tagNames),
		})
	}

	return exportRows, nil
}

func joinTagNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += "," + names[i]
	}
	return result
}

func (m *ReportManager) DefaultExportFilename(params ReportParams) (string, error) {
	filters, err := m.resolveReportFilters(params)
	if err != nil {
		return "", err
	}

	t, err := time.Parse("2006-01-02", filters.DateFrom)
	if err != nil {
		return fmt.Sprintf("wallet-report-%s.csv", filters.DateFrom), nil
	}

	return fmt.Sprintf("wallet-report-%s.csv", t.Format("2006-01")), nil
}
