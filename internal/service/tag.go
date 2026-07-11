package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (s *Service) GetTagByID(id int64) (*gen.Tag, error) {
	tag, err := s.q.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}
	return tag, nil
}

func (s *Service) GetTagByName(name string) (*gen.Tag, error) {
	tag, err := s.q.GetTagByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "tag", Name: name}
		}
		return nil, err
	}
	return tag, nil
}

func (s *Service) ListTags() ([]*gen.Tag, error) {
	return s.q.ListTags(s.ctx())
}

func (s *Service) CreateTag(name string) (*gen.Tag, error) {
	if name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "tag name is required"}
	}

	existing, err := s.q.GetTagByName(s.ctx(), name)
	if err == nil && existing != nil {
		return nil, shared.ErrDuplicateName
	}

	return s.q.CreateTag(s.ctx(), name)
}

func (s *Service) DeleteTag(id int64) error {
	_, err := s.q.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &shared.NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		return err
	}
	return s.q.DeleteTag(s.ctx(), id)
}

func (s *Service) ListTransactionTags(transactionID int64) ([]*gen.Tag, error) {
	return s.q.ListTransactionTags(s.ctx(), transactionID)
}

func (s *Service) AddTransactionTag(transactionID, tagID int64) error {
	return s.q.AddTransactionTag(s.ctx(), gen.AddTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
}

func (s *Service) RemoveTransactionTag(transactionID, tagID int64) error {
	return s.q.RemoveTransactionTag(s.ctx(), gen.RemoveTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
}
