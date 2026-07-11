package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/budget"
	"github.com/afadhitya/wallet-app/internal/service/plannedpayment"
	"github.com/afadhitya/wallet-app/internal/service/report"
	"github.com/afadhitya/wallet-app/internal/service/shared"
	"github.com/afadhitya/wallet-app/internal/service/transaction"
)

type Service struct {
	q  gen.Querier
	db *sql.DB
	*transaction.Manager
	*plannedpayment.PlannedPaymentManager
	*budget.BudgetManager
	*report.ReportManager
}

func New(database *sql.DB) *Service {
	q := gen.New(database)
	return &Service{
		q:                    q,
		db:                   database,
		Manager:              transaction.NewManager(q),
		PlannedPaymentManager: plannedpayment.NewPlannedPaymentManager(q),
		BudgetManager:        budget.NewBudgetManager(q),
		ReportManager:        report.NewReportManager(q),
	}
}

func NewWithQuerier(database *sql.DB, querier gen.Querier) *Service {
	return &Service{
		q:                    querier,
		db:                   database,
		Manager:              transaction.NewManager(querier),
		PlannedPaymentManager: plannedpayment.NewPlannedPaymentManager(querier),
		BudgetManager:        budget.NewBudgetManager(querier),
		ReportManager:        report.NewReportManager(querier),
	}
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) Queries() gen.Querier {
	return s.q
}

func (s *Service) ctx() context.Context {
	return context.Background()
}

func (s *Service) ResolveCategory(identifier string) (*gen.Category, error) {
	return shared.ResolveCategory(s.q, identifier)
}

func (s *Service) ResolveAccount(identifier string) (*gen.Account, error) {
	return shared.ResolveAccount(s.q, identifier)
}

func (s *Service) ResolveTag(identifier string) (*gen.Tag, error) {
	return shared.ResolveTag(s.q, identifier)
}

func (s *Service) GetBaseCurrency() (string, error) {
	return shared.GetBaseCurrency()
}

func (s *Service) Convert(amount int64, fromCurrency string) (int64, error) {
	return shared.Convert(amount, fromCurrency)
}

func (s *Service) GetRate(currency string) (int64, error) {
	return shared.GetRate(currency)
}

func (s *Service) ListRates() (string, map[string]int64, error) {
	return shared.ListRates()
}

func (s *Service) AddRate(currency string, rate int64) error {
	return shared.AddRate(currency, rate)
}

func (s *Service) SetRate(currency string, rate int64) error {
	return shared.SetRate(currency, rate)
}

func (s *Service) RemoveRate(currency string) error {
	return shared.RemoveRate(currency)
}

func parseDate(input string) (string, error) {
	return shared.ParseDate(input)
}

func parseMonth(input string) (string, string, error) {
	return shared.ParseMonth(input)
}

func (s *Service) GetAccountByID(id int64) (*gen.Account, error) {
	account, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "account", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}
	return account, nil
}

func (s *Service) GetAccountByName(name string) (*gen.Account, error) {
	account, err := s.q.GetAccountByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "account", Name: name}
		}
		return nil, err
	}
	return account, nil
}

type NotFoundError = shared.NotFoundError
type ValidationError = shared.ValidationError
type TestRateConfig = shared.TestRateConfig

type (
	CreateExpenseParams    = transaction.CreateExpenseParams
	CreateIncomeParams     = transaction.CreateIncomeParams
	TransactionResult      = transaction.TransactionResult
	CreateTransferParams   = transaction.CreateTransferParams
	TransferResult         = transaction.TransferResult
	ListTransactionsParams = transaction.ListTransactionsParams
	ListTransactionsResult = transaction.ListTransactionsResult
	EditTransactionParams  = transaction.EditTransactionParams
	AdjustBalanceParams    = transaction.AdjustBalanceParams
	AdjustBalanceResult    = transaction.AdjustBalanceResult

	CreatePlannedPaymentParams = plannedpayment.CreatePlannedPaymentParams
	EditPlannedPaymentParams   = plannedpayment.EditPlannedPaymentParams
	PayPlannedPaymentParams    = plannedpayment.PayPlannedPaymentParams
	PayPlannedPaymentResult    = plannedpayment.PayPlannedPaymentResult
	ListDueFilter              = plannedpayment.ListDueFilter
	ListDueParams              = plannedpayment.ListDueParams

	SetBudgetParams    = budget.SetBudgetParams
	BudgetResult       = budget.BudgetResult
	ListBudgetsParams  = budget.ListBudgetsParams
	BudgetListItem     = budget.BudgetListItem
	CheckBudgetsParams = budget.CheckBudgetsParams
	CheckBudgetResult  = budget.CheckBudgetResult
	EditBudgetParams   = budget.EditBudgetParams

	ReportParams       = report.ReportParams
	ReportFilters      = report.ReportFilters
	ReportCategoryRow  = report.ReportCategoryRow
	ReportAccountRow   = report.ReportAccountRow
	ReportTagRow       = report.ReportTagRow
	ReportExportRow    = report.ReportExportRow
	ReportResult       = report.ReportResult
)

var (
	ErrNotFound      = shared.ErrNotFound
	ErrDuplicateName = shared.ErrDuplicateName
	ErrInvalidAmount = shared.ErrInvalidAmount
	ErrMissingField  = shared.ErrMissingField

	ErrInvalidMonth  = report.ErrInvalidMonth
	ErrInvalidExport = report.ErrInvalidExport
	ErrInvalidBy     = report.ErrInvalidBy
	ErrExportFailed  = report.ErrExportFailed
	ErrNoReportData  = report.ErrNoReportData

	DueCurrentMonth = plannedpayment.DueCurrentMonth
	DueCurrentWeek  = plannedpayment.DueCurrentWeek
	DueOverdue      = plannedpayment.DueOverdue
	DueNextDays     = plannedpayment.DueNextDays

	BudgetStatusOK      = budget.BudgetStatusOK
	BudgetStatusWarning = budget.BudgetStatusWarning
	BudgetStatusOver    = budget.BudgetStatusOver
)
