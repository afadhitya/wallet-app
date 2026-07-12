package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/afadhitya/wallet-app/internal/gen"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrDuplicateName = errors.New("name already exists")
	ErrInvalidAmount = errors.New("amount must be positive")
	ErrMissingField  = errors.New("required field missing")
)

type NotFoundError struct {
	Entity string
	Name   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Entity, e.Name)
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Service struct {
	q      gen.Querier
	db     *sql.DB
	logger *slog.Logger
}

func New(database *sql.DB, logger *slog.Logger) *Service {
	return &Service{
		q:      gen.New(database),
		db:     database,
		logger: logger,
	}
}

func NewWithQuerier(database *sql.DB, querier gen.Querier, logger *slog.Logger) *Service {
	return &Service{
		q:      querier,
		db:     database,
		logger: logger,
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

func isBusinessError(err error) bool {
	var notFound *NotFoundError
	var validation *ValidationError
	var rateNotFound *RateNotFoundError
	if errors.As(err, &notFound) || errors.As(err, &validation) || errors.As(err, &rateNotFound) {
		return true
	}
	if errors.Is(err, ErrInvalidAmount) || errors.Is(err, ErrDuplicateName) || errors.Is(err, ErrMissingField) ||
		errors.Is(err, ErrRateConfigMissing) || errors.Is(err, ErrRateMustBePositive) {
		return true
	}
	return false
}
