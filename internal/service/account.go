package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (s *Service) ListAccounts() ([]*gen.Account, error) {
	return s.q.ListAccounts(s.ctx())
}

func (s *Service) ListAllAccounts() ([]*gen.Account, error) {
	return s.q.ListAllAccounts(s.ctx())
}

func (s *Service) CreateAccount(name, accountType, currency string) (*gen.Account, error) {
	if name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "account name is required"}
	}
	if accountType == "" {
		accountType = "checking"
	}
	if currency == "" {
		currency = "IDR"
	}

	existing, checkErr := s.q.GetAccountByName(s.ctx(), name)
	if checkErr == nil && existing != nil {
		return nil, shared.ErrDuplicateName
	}
	if checkErr != nil && !errors.Is(checkErr, sql.ErrNoRows) {
		return nil, checkErr
	}

	return s.q.CreateAccount(s.ctx(), gen.CreateAccountParams{
		Name:     name,
		Type:     accountType,
		Currency: currency,
	})
}

func (s *Service) UpdateAccount(id int64, name, accountType, currency string, sortOrder int64) (*gen.Account, error) {
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}

	if name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "account name cannot be empty"}
	}

	return s.q.UpdateAccount(s.ctx(), gen.UpdateAccountParams{
		ID:        id,
		Name:      name,
		Type:      accountType,
		Currency:  currency,
		SortOrder: sortOrder,
	})
}

func (s *Service) ArchiveAccount(id int64) error {
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &shared.NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		return err
	}
	return s.q.ArchiveAccount(s.ctx(), id)
}

func (s *Service) UpdateAccountBalance(id int64, balance int64) error {
	return s.q.UpdateAccountBalance(s.ctx(), gen.UpdateAccountBalanceParams{
		ID:      id,
		Balance: balance,
	})
}
