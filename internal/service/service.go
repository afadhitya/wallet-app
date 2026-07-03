package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	queries *gen.Queries
	db      *sql.DB
}

func New(database *sql.DB) *Service {
	return &Service{
		queries: gen.New(database),
		db:      database,
	}
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) Queries() *gen.Queries {
	return s.queries
}

func (s *Service) ctx() context.Context {
	return context.Background()
}
