package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func (s *Service) GetCategoryByID(id int64) (*gen.Category, error) {
	category, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}
	return category, nil
}

func (s *Service) ListCategories() ([]*gen.Category, error) {
	return s.q.ListCategories(s.ctx())
}

func (s *Service) ListAllCategories() ([]*gen.Category, error) {
	return s.q.ListAllCategories(s.ctx())
}

func (s *Service) CreateCategory(name, parentIDStr, icon string) (*gen.Category, error) {
	if name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "category name is required"}
	}

	var parentID sql.NullInt64
	if parentIDStr != "" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			return nil, &shared.ValidationError{Field: "parent_id", Message: "invalid parent category ID"}
		}
		parentID = sql.NullInt64{Int64: id, Valid: true}

		_, err = s.q.GetCategoryByID(s.ctx(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, &shared.NotFoundError{Entity: "parent category", Name: parentIDStr}
			}
			return nil, err
		}
	}

	existing, err := s.q.GetCategoryByName(s.ctx(), name)
	if err == nil && existing != nil {
		return nil, shared.ErrDuplicateName
	}

	var iconVal sql.NullString
	if icon != "" {
		iconVal = sql.NullString{String: icon, Valid: true}
	}

	return s.q.CreateCategory(s.ctx(), gen.CreateCategoryParams{
		Name:     name,
		ParentID: parentID,
		Type:     "expense",
		Icon:     iconVal,
	})
}

func (s *Service) UpdateCategory(id int64, name, icon string) (*gen.Category, error) {
	_, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		return nil, err
	}

	if name == "" {
		return nil, &shared.ValidationError{Field: "name", Message: "category name cannot be empty"}
	}

	var iconVal sql.NullString
	if icon != "" {
		iconVal = sql.NullString{String: icon, Valid: true}
	}

	return s.q.UpdateCategory(s.ctx(), gen.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Icon: iconVal,
	})
}

func (s *Service) ArchiveCategory(id int64) error {
	_, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &shared.NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		return err
	}
	return s.q.ArchiveCategory(s.ctx(), id)
}
