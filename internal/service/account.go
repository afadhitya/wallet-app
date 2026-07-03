package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (s *Service) GetAccountByID(id int64) (*gen.Account, error) {
	account, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}
	return account, nil
}

func (s *Service) GetAccountByName(name string) (*gen.Account, error) {
	account, err := s.q.GetAccountByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "account", Name: name}
		}
		return nil, err
	}
	return account, nil
}

func (s *Service) ResolveAccount(identifier string) (*gen.Account, error) {
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		account, err := s.q.GetAccountByID(s.ctx(), id)
		if err == nil {
			return account, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	account, err := s.q.GetAccountByName(s.ctx(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "account", Name: identifier}
		}
		return nil, err
	}
	return account, nil
}

func (s *Service) ListAccounts() ([]*gen.Account, error) {
	return s.q.ListAccounts(s.ctx())
}

func (s *Service) CreateAccount(name, accountType, currency string) (*gen.Account, error) {
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "account name is required"}
	}
	if accountType == "" {
		accountType = "checking"
	}
	if currency == "" {
		currency = "IDR"
	}

	return s.q.CreateAccount(s.ctx(), gen.CreateAccountParams{
		Name:     name,
		Type:     accountType,
		Currency: currency,
	})
}

func (s *Service) UpdateAccount(id int64, name, accountType, currency string) (*gen.Account, error) {
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}

	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "account name cannot be empty"}
	}

	return s.q.UpdateAccount(s.ctx(), gen.UpdateAccountParams{
		ID:       id,
		Name:     name,
		Type:     accountType,
		Currency: currency,
	})
}

func (s *Service) ArchiveAccount(id int64) error {
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
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
