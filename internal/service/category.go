package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func (s *Service) GetCategoryByID(id int64) (*gen.Category, error) {
	s.logger.Info("GetCategoryByID called", slog.Int64("id", id))
	category, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("GetCategoryByID not found", slog.Int64("id", id))
			return nil, &NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("GetCategoryByID failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("GetCategoryByID completed", slog.Int64("id", id))
	return category, nil
}

func (s *Service) ResolveCategory(identifier string) (*gen.Category, error) {
	s.logger.Info("ResolveCategory called", slog.String("identifier", identifier))
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		category, err := s.q.GetCategoryByID(s.ctx(), id)
		if err == nil {
			s.logger.Info("ResolveCategory completed", slog.String("identifier", identifier))
			return category, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("ResolveCategory failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

	category, err := s.q.GetCategoryByName(s.ctx(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			suggestions, _ := s.q.GetCategorySuggestions(s.ctx(), fmt.Sprintf("%%%s%%", identifier))
			if len(suggestions) > 0 {
				var names []string
				for _, sug := range suggestions {
					names = append(names, sug.Name)
				}
				errMsg := fmt.Errorf("category '%s' not found. Did you mean: %v?", identifier, names)
				s.logger.Warn("ResolveCategory not found", slog.String("identifier", identifier))
				return nil, errMsg
			}
			s.logger.Warn("ResolveCategory not found", slog.String("identifier", identifier))
			return nil, &NotFoundError{Entity: "category", Name: identifier}
		}
		s.logger.Error("ResolveCategory failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ResolveCategory completed", slog.String("identifier", identifier))
	return category, nil
}

func (s *Service) ListCategories() ([]*gen.Category, error) {
	s.logger.Info("ListCategories called")
	categories, err := s.q.ListCategories(s.ctx())
	if err != nil {
		s.logger.Error("ListCategories failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListCategories completed", slog.Int("count", len(categories)))
	return categories, nil
}

func (s *Service) ListAllCategories() ([]*gen.Category, error) {
	s.logger.Info("ListAllCategories called")
	categories, err := s.q.ListAllCategories(s.ctx())
	if err != nil {
		s.logger.Error("ListAllCategories failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("ListAllCategories completed", slog.Int("count", len(categories)))
	return categories, nil
}

func (s *Service) CreateCategory(name, parentIDStr, icon string) (*gen.Category, error) {
	s.logger.Info("CreateCategory called", slog.String("name", name), slog.String("parent_id", parentIDStr), slog.String("icon", icon))
	if name == "" {
		s.logger.Warn("CreateCategory validation failed", slog.String("field", "name"))
		return nil, &ValidationError{Field: "name", Message: "category name is required"}
	}

	var parentID sql.NullInt64
	if parentIDStr != "" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			s.logger.Warn("CreateCategory validation failed", slog.String("field", "parent_id"))
			return nil, &ValidationError{Field: "parent_id", Message: "invalid parent category ID"}
		}
		parentID = sql.NullInt64{Int64: id, Valid: true}

		_, err = s.q.GetCategoryByID(s.ctx(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("CreateCategory parent not found", slog.String("parent_id", parentIDStr))
				return nil, &NotFoundError{Entity: "parent category", Name: parentIDStr}
			}
			s.logger.Error("CreateCategory failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

	existing, err := s.q.GetCategoryByName(s.ctx(), name)
	if err == nil && existing != nil {
		s.logger.Warn("CreateCategory duplicate name", slog.String("name", name))
		return nil, ErrDuplicateName
	}

	var iconVal sql.NullString
	if icon != "" {
		iconVal = sql.NullString{String: icon, Valid: true}
	}

	result, err := s.q.CreateCategory(s.ctx(), gen.CreateCategoryParams{
		Name:     name,
		ParentID: parentID,
		Type:     "expense",
		Icon:     iconVal,
	})
	if err != nil {
		s.logger.Error("CreateCategory failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("CreateCategory completed", slog.Int64("id", result.ID), slog.String("name", result.Name))
	return result, nil
}

func (s *Service) UpdateCategory(id int64, name, icon string) (*gen.Category, error) {
	s.logger.Info("UpdateCategory called", slog.Int64("id", id), slog.String("name", name), slog.String("icon", icon))
	_, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("UpdateCategory not found", slog.Int64("id", id))
			return nil, &NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("UpdateCategory failed", slog.String("error", err.Error()))
		return nil, err
	}

	if name == "" {
		s.logger.Warn("UpdateCategory validation failed", slog.String("field", "name"))
		return nil, &ValidationError{Field: "name", Message: "category name cannot be empty"}
	}

	var iconVal sql.NullString
	if icon != "" {
		iconVal = sql.NullString{String: icon, Valid: true}
	}

	result, err := s.q.UpdateCategory(s.ctx(), gen.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Icon: iconVal,
	})
	if err != nil {
		s.logger.Error("UpdateCategory failed", slog.String("error", err.Error()))
		return nil, err
	}
	s.logger.Info("UpdateCategory completed", slog.Int64("id", id))
	return result, nil
}

func (s *Service) ArchiveCategory(id int64) error {
	s.logger.Info("ArchiveCategory called", slog.Int64("id", id))
	_, err := s.q.GetCategoryByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("ArchiveCategory not found", slog.Int64("id", id))
			return &NotFoundError{Entity: "category", Name: strconv.FormatInt(id, 10)}
		}
		s.logger.Error("ArchiveCategory failed", slog.String("error", err.Error()))
		return err
	}
	err = s.q.ArchiveCategory(s.ctx(), id)
	if err != nil {
		s.logger.Error("ArchiveCategory failed", slog.String("error", err.Error()))
		return err
	}
	s.logger.Info("ArchiveCategory completed", slog.Int64("id", id))
	return nil
}
