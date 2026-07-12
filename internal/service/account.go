package service

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (s *Service) GetAccountByID(id int64) (*gen.Account, error) {
	s.logger.Info("GetAccountByID called", slog.Int64("id", id))
	account, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetAccountByID not found", slog.Int64("id", id))
			return nil, &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("GetAccountByID failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetAccountByID completed", slog.Int64("id", id))
	return account, nil
}

func (s *Service) GetAccountByName(name string) (*gen.Account, error) {
	s.logger.Info("GetAccountByName called", slog.String("name", name))
	account, err := s.q.GetAccountByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetAccountByName not found", slog.String("name", name))
			return nil, &NotFoundError{Entity: "account", Name: name}
		}
		s.logger.Error("GetAccountByName failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetAccountByName completed", slog.String("name", name))
	return account, nil
}

func (s *Service) ResolveAccount(identifier string) (*gen.Account, error) {
	s.logger.Info("ResolveAccount called", slog.String("identifier", identifier))
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		account, err := s.q.GetAccountByID(s.ctx(), id)
		if err == nil {
			s.logger.Info("ResolveAccount completed", slog.String("identifier", identifier))
			return account, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("ResolveAccount failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

	account, err := s.q.GetAccountByName(s.ctx(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("ResolveAccount not found", slog.String("identifier", identifier))
			return nil, &NotFoundError{Entity: "account", Name: identifier}
		}
		s.logger.Error("ResolveAccount failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ResolveAccount completed", slog.String("identifier", identifier))
	return account, nil
}

func (s *Service) ListAccounts() ([]*gen.Account, error) {
	s.logger.Info("ListAccounts called")
	accounts, err := s.q.ListAccounts(s.ctx())
	if err != nil {
		s.logger.Error("ListAccounts failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListAccounts completed", slog.Int("count", len(accounts)))
	return accounts, nil
}

func (s *Service) ListAllAccounts() ([]*gen.Account, error) {
	s.logger.Info("ListAllAccounts called")
	accounts, err := s.q.ListAllAccounts(s.ctx())
	if err != nil {
		s.logger.Error("ListAllAccounts failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListAllAccounts completed", slog.Int("count", len(accounts)))
	return accounts, nil
}

func (s *Service) CreateAccount(name, accountType, currency string) (*gen.Account, error) {
	s.logger.Info("CreateAccount called", slog.String("name", name), slog.String("type", accountType), slog.String("currency", currency))
	if name == "" {
		s.logger.Warn("CreateAccount validation failed", slog.String("field", "name"))
		return nil, &ValidationError{Field: "name", Message: "account name is required"}
	}
	if accountType == "" {
		accountType = "checking"
	}
	if currency == "" {
		currency = "IDR"
	}

	existing, checkErr := s.q.GetAccountByName(s.ctx(), name)
	if checkErr == nil && existing != nil {
		s.logger.Warn("CreateAccount duplicate name", slog.String("name", name))
		return nil, ErrDuplicateName
	}
	if checkErr != nil && !errors.Is(checkErr, sql.ErrNoRows) {
		s.logger.Error("CreateAccount failed checking duplicate", slog.String("error", checkErr.Error()))
		return nil, checkErr
	}

	result, err := s.q.CreateAccount(s.ctx(), gen.CreateAccountParams{
		Name:     name,
		Type:     accountType,
		Currency: currency,
	})
	if err != nil {
		s.logger.Error("CreateAccount failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("CreateAccount completed", slog.Int64("id", result.ID))
	return result, nil
}

func (s *Service) UpdateAccount(id int64, name, accountType, currency string, sortOrder int64) (*gen.Account, error) {
	s.logger.Info("UpdateAccount called", slog.Int64("id", id), slog.String("name", name))
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("UpdateAccount not found", slog.Int64("id", id))
			return nil, &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("UpdateAccount failed", slog.String("error", err.Error()))
		return nil, err
	}

	if name == "" {
		s.logger.Warn("UpdateAccount validation failed", slog.String("field", "name"))
		return nil, &ValidationError{Field: "name", Message: "account name cannot be empty"}
	}

	result, err := s.q.UpdateAccount(s.ctx(), gen.UpdateAccountParams{
		ID:        id,
		Name:      name,
		Type:      accountType,
		Currency:  currency,
		SortOrder: sortOrder,
	})
	if err != nil {
		s.logger.Error("UpdateAccount failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("UpdateAccount completed", slog.Int64("id", id))
	return result, nil
}

func (s *Service) ArchiveAccount(id int64) error {
	s.logger.Info("ArchiveAccount called", slog.Int64("id", id))
	_, err := s.q.GetAccountByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("ArchiveAccount not found", slog.Int64("id", id))
			return &NotFoundError{Entity: "account", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("ArchiveAccount failed", slog.String("error", err.Error()))
		return err
	}
	err = s.q.ArchiveAccount(s.ctx(), id)
	if err != nil {
		s.logger.Error("ArchiveAccount failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("ArchiveAccount completed", slog.Int64("id", id))
	return nil
}

func (s *Service) UpdateAccountBalance(id int64, balance int64) error {
	s.logger.Info("UpdateAccountBalance called", slog.Int64("id", id), slog.Int64("balance", balance))
	err := s.q.UpdateAccountBalance(s.ctx(), gen.UpdateAccountBalanceParams{
		ID:      id,
		Balance: balance,
	})
	if err != nil {
		s.logger.Error("UpdateAccountBalance failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("UpdateAccountBalance completed", slog.Int64("id", id))
	return nil
}
