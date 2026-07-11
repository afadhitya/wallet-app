package report

import "github.com/afadhitya/wallet-app/internal/gen"

type ReportManager struct {
	q gen.Querier
}

func NewReportManager(q gen.Querier) *ReportManager {
	return &ReportManager{q: q}
}
