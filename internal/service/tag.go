package service

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (s *Service) GetTagByID(id int64) (*gen.Tag, error) {
	s.logger.Info("GetTagByID called", slog.Int64("id", id))
	tag, err := s.q.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetTagByID not found", slog.Int64("id", id))
			return nil, &NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("GetTagByID failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetTagByID completed", slog.Int64("id", id))
	return tag, nil
}

func (s *Service) GetTagByName(name string) (*gen.Tag, error) {
	s.logger.Info("GetTagByName called", slog.String("name", name))
	tag, err := s.q.GetTagByName(s.ctx(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetTagByName not found", slog.String("name", name))
			return nil, &NotFoundError{Entity: "tag", Name: name}
		}
		s.logger.Error("GetTagByName failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetTagByName completed", slog.String("name", name))
	return tag, nil
}

func (s *Service) ResolveTag(identifier string) (*gen.Tag, error) {
	s.logger.Info("ResolveTag called", slog.String("identifier", identifier))
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		tag, err := s.q.GetTagByID(s.ctx(), id)
		if err == nil {
			s.logger.Info("ResolveTag completed", slog.String("identifier", identifier))
			return tag, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("ResolveTag failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

	tag, err := s.q.GetTagByName(s.ctx(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("ResolveTag not found", slog.String("identifier", identifier))
			return nil, &NotFoundError{Entity: "tag", Name: identifier}
		}
		s.logger.Error("ResolveTag failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ResolveTag completed", slog.String("identifier", identifier))
	return tag, nil
}

func (s *Service) ListTags() ([]*gen.Tag, error) {
	s.logger.Info("ListTags called")
	tags, err := s.q.ListTags(s.ctx())
	if err != nil {
		s.logger.Error("ListTags failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListTags completed", slog.Int("count", len(tags)))
	return tags, nil
}

func (s *Service) CreateTag(name string) (*gen.Tag, error) {
	s.logger.Info("CreateTag called", slog.String("name", name))
	if name == "" {
		s.logger.Warn("CreateTag validation failed", slog.String("field", "name"))
		return nil, &ValidationError{Field: "name", Message: "tag name is required"}
	}

	existing, err := s.q.GetTagByName(s.ctx(), name)
	if err == nil && existing != nil {
		s.logger.Warn("CreateTag duplicate name", slog.String("name", name))
		return nil, ErrDuplicateName
	}

	result, err := s.q.CreateTag(s.ctx(), name)
	if err != nil {
		s.logger.Error("CreateTag failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("CreateTag completed", slog.Int64("id", result.ID), slog.String("name", result.Name))
	return result, nil
}

func (s *Service) DeleteTag(id int64) error {
	s.logger.Info("DeleteTag called", slog.Int64("id", id))
	_, err := s.q.GetTagByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("DeleteTag not found", slog.Int64("id", id))
			return &NotFoundError{Entity: "tag", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("DeleteTag failed", slog.String("error", err.Error()))
		return err
	}
	err = s.q.DeleteTag(s.ctx(), id)
	if err != nil {
		s.logger.Error("DeleteTag failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("DeleteTag completed", slog.Int64("id", id))
	return nil
}

func (s *Service) ListTransactionTags(transactionID int64) ([]*gen.Tag, error) {
	s.logger.Info("ListTransactionTags called", slog.Int64("transaction_id", transactionID))
	tags, err := s.q.ListTransactionTags(s.ctx(), transactionID)
	if err != nil {
		s.logger.Error("ListTransactionTags failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListTransactionTags completed", slog.Int("count", len(tags)))
	return tags, nil
}

func (s *Service) AddTransactionTag(transactionID, tagID int64) error {
	s.logger.Info("AddTransactionTag called", slog.Int64("transaction_id", transactionID), slog.Int64("tag_id", tagID))
	err := s.q.AddTransactionTag(s.ctx(), gen.AddTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
	if err != nil {
		s.logger.Error("AddTransactionTag failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("AddTransactionTag completed", slog.Int64("transaction_id", transactionID), slog.Int64("tag_id", tagID))
	return nil
}

func (s *Service) RemoveTransactionTag(transactionID, tagID int64) error {
	s.logger.Info("RemoveTransactionTag called", slog.Int64("transaction_id", transactionID), slog.Int64("tag_id", tagID))
	err := s.q.RemoveTransactionTag(s.ctx(), gen.RemoveTransactionTagParams{
		TransactionID: transactionID,
		TagID:         tagID,
	})
	if err != nil {
		s.logger.Error("RemoveTransactionTag failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("RemoveTransactionTag completed", slog.Int64("transaction_id", transactionID), slog.Int64("tag_id", tagID))
	return nil
}
