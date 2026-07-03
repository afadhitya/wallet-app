package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (s *Service) GetTagByID(id int64) (*gen.Tag, error) {
	tag, err := s.queries.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}
	return tag, nil
}

func (s *Service) GetTagByName(name string) (*gen.Tag, error) {
	tag, err := s.queries.GetTagByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "tag", Name: name}
		}
		return nil, err
	}
	return tag, nil
}

func (s *Service) ResolveTag(identifier string) (*gen.Tag, error) {
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		tag, err := s.queries.GetTagByID(s.ctx(), id)
		if err == nil {
			return tag, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	tag, err := s.queries.GetTagByName(s.ctx(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "tag", Name: identifier}
		}
		return nil, err
	}
	return tag, nil
}

func (s *Service) ListTags() ([]*gen.Tag, error) {
	return s.queries.ListTags(s.ctx())
}

func (s *Service) CreateTag(name string) (*gen.Tag, error) {
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "tag name is required"}
	}

	existing, err := s.queries.GetTagByName(s.ctx(), name)
	if err == nil && existing != nil {
		return nil, ErrDuplicateName
	}

	return s.queries.CreateTag(s.ctx(), name)
}

func (s *Service) DeleteTag(id int64) error {
	_, err := s.queries.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		return err
	}
	return s.queries.DeleteTag(s.ctx(), id)
}

func (s *Service) ListTransactionTags(transactionID int64) ([]*gen.Tag, error) {
	return s.queries.ListTransactionTags(s.ctx(), transactionID)
}

func (s *Service) AddTransactionTag(transactionID, tagID int64) error {
	return s.queries.AddTransactionTag(s.ctx(), gen.AddTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
}

func (s *Service) RemoveTransactionTag(transactionID, tagID int64) error {
	return s.queries.RemoveTransactionTag(s.ctx(), gen.RemoveTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
}
