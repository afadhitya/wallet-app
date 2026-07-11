package shared

import (
	"errors"
	"fmt"
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
