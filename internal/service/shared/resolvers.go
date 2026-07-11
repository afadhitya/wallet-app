package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/afadhitya/wallet-app/internal/gen"
)

func ResolveCategory(q gen.Querier, identifier string) (*gen.Category, error) {
	ctx := context.Background()
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		category, err := q.GetCategoryByID(ctx, id)
		if err == nil {
			return category, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	category, err := q.GetCategoryByName(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			suggestions, _ := q.GetCategorySuggestions(ctx, fmt.Sprintf("%%%s%%", identifier))
			if len(suggestions) > 0 {
				var names []string
				for _, sug := range suggestions {
					names = append(names, sug.Name)
				}
				return nil, fmt.Errorf("category '%s' not found. Did you mean: %v?", identifier, names)
			}
			return nil, &NotFoundError{Entity: "category", Name: identifier}
		}
		return nil, err
	}
	return category, nil
}

func ResolveAccount(q gen.Querier, identifier string) (*gen.Account, error) {
	ctx := context.Background()
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		account, err := q.GetAccountByID(ctx, id)
		if err == nil {
			return account, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	account, err := q.GetAccountByName(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "account", Name: identifier}
		}
		return nil, err
	}
	return account, nil
}

func ResolveTag(q gen.Querier, identifier string) (*gen.Tag, error) {
	ctx := context.Background()
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		tag, err := q.GetTagByID(ctx, id)
		if err == nil {
			return tag, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	tag, err := q.GetTagByName(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "tag", Name: identifier}
		}
		return nil, err
	}
	return tag, nil
}
