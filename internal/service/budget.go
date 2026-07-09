package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/afadhitya/wallet-app/internal/gen"
)

type SetBudgetParams struct {
	Name          string
	Amount        int64
	Period        string
	From          string
	To            string
	NotifyPct     int64
	Categories    []string
	AllCategories bool
	Tags          []string
}

type BudgetResult struct {
	Budget     *gen.Budget
	Categories []*gen.Category
	Tags       []*gen.Tag
}

type ListBudgetsParams struct {
	All bool
}

type BudgetListItem struct {
	Budget     *gen.Budget
	Spent      int64
	Remaining  int64
	Categories []*gen.Category
	Tags       []*gen.Tag
}

type CheckBudgetsParams struct {
	Identifier string
	All        bool
}

type CheckBudgetResult struct {
	Budget      *gen.Budget
	Spent       int64
	Remaining   int64
	PercentUsed float64
	Status      string
	Categories  []*gen.Category
	Tags        []*gen.Tag
}

type EditBudgetParams struct {
	Amount           *int64
	Name             string
	NotifyPct        *int64
	AddCategories    []string
	RemoveCategories []string
	AddTags          []string
	RemoveTags       []string
}

const (
	BudgetStatusOK      = "ok"
	BudgetStatusWarning = "warning"
	BudgetStatusOver    = "over"
)

var validPeriods = map[string]bool{
	"monthly":  true,
	"weekly":   true,
	"yearly":   true,
	"one_time": true,
}

func calculatePeriod(periodType, from, to string) (string, string, error) {
	now := time.Now()

	if from != "" && to != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return "", "", &ValidationError{Field: "from", Message: "invalid date format (use YYYY-MM-DD)"}
		}
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return "", "", &ValidationError{Field: "to", Message: "invalid date format (use YYYY-MM-DD)"}
		}
		return from, to, nil
	}

	switch periodType {
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	case "weekly":
		weekday := now.Weekday()
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 6)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
	default:
		return "", "", &ValidationError{Field: "period", Message: "one_time budget requires --from and --to dates"}
	}
}

func (s *Service) SetBudget(params SetBudgetParams) (*BudgetResult, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if len(params.Categories) == 0 && len(params.Tags) == 0 && !params.AllCategories {
		return nil, &ValidationError{Field: "targets", Message: "budget must have at least one category or tag target"}
	}

	if !validPeriods[params.Period] {
		return nil, &ValidationError{Field: "period", Message: "supported periods: monthly, weekly, yearly, one_time"}
	}

	if params.NotifyPct == 0 {
		params.NotifyPct = 80
	}
	if params.NotifyPct < 1 || params.NotifyPct > 100 {
		return nil, &ValidationError{Field: "notify", Message: "notification threshold must be between 1 and 100"}
	}

	periodStart, periodEnd, err := calculatePeriod(params.Period, params.From, params.To)
	if err != nil {
		return nil, err
	}

	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}

	resolvedCategories := make([]*gen.Category, 0, len(params.Categories))
	for _, catName := range params.Categories {
		cat, err := s.ResolveCategory(catName)
		if err != nil {
			return nil, fmt.Errorf("category '%s': %w", catName, err)
		}
		resolvedCategories = append(resolvedCategories, cat)
	}

	resolvedTags := make([]*gen.Tag, 0, len(params.Tags))
	for _, tagName := range params.Tags {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("tag '%s': %w", tagName, err)
		}
		resolvedTags = append(resolvedTags, tag)
	}

	existing, err := s.q.GetBudgetByNameAndPeriod(s.ctx(), gen.GetBudgetByNameAndPeriodParams{
		Name:        name,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing budget: %w", err)
	}
	if err == nil && existing != nil {
		return s.updateExistingBudget(existing.ID, params, periodTypeForDB(params.Period), periodStart, periodEnd, resolvedCategories, resolvedTags)
	}

	notifyPct := sql.NullInt64{Int64: params.NotifyPct, Valid: true}

	budget, err := s.q.CreateBudget(s.ctx(), gen.CreateBudgetParams{
		Name:          name,
		Amount:        params.Amount,
		Currency:      "IDR",
		Type:          periodTypeForDB(params.Period),
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		NotifyAtPct:   notifyPct,
		AllCategories: boolToInt64(params.AllCategories),
	})
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	for _, cat := range resolvedCategories {
		if err := s.q.AddBudgetCategory(s.ctx(), gen.AddBudgetCategoryParams{
			BudgetID:   budget.ID,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, tag := range resolvedTags {
		if err := s.q.AddBudgetTag(s.ctx(), gen.AddBudgetTagParams{
			BudgetID: budget.ID,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	return &BudgetResult{
		Budget:     budget,
		Categories: resolvedCategories,
		Tags:       resolvedTags,
	}, nil
}

func (s *Service) updateExistingBudget(id int64, params SetBudgetParams, periodType, periodStart, periodEnd string, categories []*gen.Category, tags []*gen.Tag) (*BudgetResult, error) {
	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}
	amountVal := sql.NullInt64{Int64: params.Amount, Valid: true}
	notifyVal := sql.NullInt64{Int64: params.NotifyPct, Valid: true}
	allCategoriesVal := sql.NullInt64{Int64: boolToInt64(params.AllCategories), Valid: true}

	budget, err := s.q.UpdateBudget(s.ctx(), gen.UpdateBudgetParams{
		ID:            id,
		Name:          name,
		Amount:        amountVal,
		NotifyAtPct:   notifyVal,
		PeriodStart:   sql.NullString{String: periodStart, Valid: true},
		PeriodEnd:     sql.NullString{String: periodEnd, Valid: true},
		Type:          sql.NullString{String: periodType, Valid: true},
		AllCategories: allCategoriesVal,
	})
	if err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}

	if err := s.q.RemoveAllBudgetCategories(s.ctx(), id); err != nil {
		return nil, fmt.Errorf("remove budget categories: %w", err)
	}
	if err := s.q.RemoveAllBudgetTags(s.ctx(), id); err != nil {
		return nil, fmt.Errorf("remove budget tags: %w", err)
	}

	for _, cat := range categories {
		if err := s.q.AddBudgetCategory(s.ctx(), gen.AddBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, tag := range tags {
		if err := s.q.AddBudgetTag(s.ctx(), gen.AddBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	return &BudgetResult{
		Budget:     budget,
		Categories: categories,
		Tags:       tags,
	}, nil
}

func periodTypeForDB(periodType string) string {
	if periodType == "one_time" {
		return "one_time"
	}
	return periodType
}

func (s *Service) ListBudgets(params ListBudgetsParams) ([]*BudgetListItem, error) {
	var budgets []*gen.Budget
	var err error

	if params.All {
		budgets, err = s.q.ListAllBudgets(s.ctx())
	} else {
		budgets, err = s.q.ListActiveBudgets(s.ctx())
	}
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	items := make([]*BudgetListItem, 0, len(budgets))
	for _, b := range budgets {
		spent, err := s.calculateSpending(b)
		if err != nil {
			return nil, fmt.Errorf("calculate spending for budget %d: %w", b.ID, err)
		}
		remaining := b.Amount - spent

		categories, _ := s.q.ListBudgetCategories(s.ctx(), b.ID)
		tags, _ := s.q.ListBudgetTags(s.ctx(), b.ID)

		items = append(items, &BudgetListItem{
			Budget:     b,
			Spent:      spent,
			Remaining:  remaining,
			Categories: categories,
			Tags:       tags,
		})
	}
	return items, nil
}

func (s *Service) CheckBudgets(params CheckBudgetsParams) ([]*CheckBudgetResult, error) {
	if params.Identifier != "" {
		return s.checkSingleBudget(params.Identifier)
	}

	budgets, err := s.q.ListActiveBudgets(s.ctx())
	if err != nil {
		return nil, fmt.Errorf("list active budgets: %w", err)
	}

	results := make([]*CheckBudgetResult, 0, len(budgets))
	for _, b := range budgets {
		current, err := s.ensureCurrentPeriod(b)
		if err != nil {
			return nil, fmt.Errorf("ensure current period for budget '%s': %w", budgetName(b), err)
		}
		result, err := s.buildCheckResult(current)
		if err != nil {
			return nil, fmt.Errorf("build check result: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) checkSingleBudget(identifier string) ([]*CheckBudgetResult, error) {
	budget, err := s.resolveBudget(identifier)
	if err != nil {
		return nil, err
	}

	current, err := s.ensureCurrentPeriod(budget)
	if err != nil {
		return nil, fmt.Errorf("ensure current period: %w", err)
	}

	result, err := s.buildCheckResult(current)
	if err != nil {
		return nil, fmt.Errorf("build check result: %w", err)
	}
	return []*CheckBudgetResult{result}, nil
}

func (s *Service) resolveBudget(identifier string) (*gen.Budget, error) {
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		budget, err := s.q.GetBudgetByID(s.ctx(), id)
		if err == nil && budget.IsActive == 1 {
			return budget, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	budgets, err := s.q.ListActiveBudgets(s.ctx())
	if err != nil {
		return nil, err
	}
	for _, b := range budgets {
		if b.Name.Valid && b.Name.String == identifier {
			return b, nil
		}
	}
	return nil, &NotFoundError{Entity: "budget", Name: identifier}
}

func (s *Service) ensureCurrentPeriod(budget *gen.Budget) (*gen.Budget, error) {
	if budget.Type == "one_time" {
		return budget, nil
	}

	now := time.Now()
	var currentStart, currentEnd string

	switch budget.Type {
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	case "weekly":
		weekday := now.Weekday()
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 6)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		currentStart = start.Format("2006-01-02")
		currentEnd = end.Format("2006-01-02")
	default:
		return budget, nil
	}

	name := budget.Name
	current, err := s.q.GetBudgetByNameAndPeriod(s.ctx(), gen.GetBudgetByNameAndPeriodParams{
		Name:        name,
		PeriodStart: currentStart,
		PeriodEnd:   currentEnd,
	})
	if err == nil && current != nil {
		return current, nil
	}

	prior, err := s.q.GetMostRecentPriorBudget(s.ctx(), gen.GetMostRecentPriorBudgetParams{
		Name:      name,
		PeriodEnd: currentStart,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return budget, nil
		}
		return nil, fmt.Errorf("get prior budget: %w", err)
	}

	newBudget, err := s.q.CreateBudget(s.ctx(), gen.CreateBudgetParams{
		Name:        prior.Name,
		Amount:      prior.Amount,
		Currency:    prior.Currency,
		Type:        prior.Type,
		PeriodStart: currentStart,
		PeriodEnd:   currentEnd,
		NotifyAtPct: prior.NotifyAtPct,
	})
	if err != nil {
		return nil, fmt.Errorf("create recurring budget: %w", err)
	}

	categories, err := s.q.ListBudgetCategories(s.ctx(), prior.ID)
	if err != nil {
		return nil, fmt.Errorf("list prior categories: %w", err)
	}
	for _, cat := range categories {
		if err := s.q.AddBudgetCategory(s.ctx(), gen.AddBudgetCategoryParams{
			BudgetID:   newBudget.ID,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("copy budget category: %w", err)
		}
	}

	tags, err := s.q.ListBudgetTags(s.ctx(), prior.ID)
	if err != nil {
		return nil, fmt.Errorf("list prior tags: %w", err)
	}
	for _, tag := range tags {
		if err := s.q.AddBudgetTag(s.ctx(), gen.AddBudgetTagParams{
			BudgetID: newBudget.ID,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("copy budget tag: %w", err)
		}
	}

	return newBudget, nil
}

func (s *Service) buildCheckResult(budget *gen.Budget) (*CheckBudgetResult, error) {
	spent, err := s.calculateSpending(budget)
	if err != nil {
		return nil, err
	}
	remaining := budget.Amount - spent

	var percentUsed float64
	if budget.Amount > 0 {
		percentUsed = float64(spent) / float64(budget.Amount) * 100
	}

	notifyPct := int64(80)
	if budget.NotifyAtPct.Valid {
		notifyPct = budget.NotifyAtPct.Int64
	}

	status := BudgetStatusOK
	if percentUsed >= 100 {
		status = BudgetStatusOver
	} else if percentUsed >= float64(notifyPct) {
		status = BudgetStatusWarning
	}

	categories, _ := s.q.ListBudgetCategories(s.ctx(), budget.ID)
	tags, _ := s.q.ListBudgetTags(s.ctx(), budget.ID)

	return &CheckBudgetResult{
		Budget:      budget,
		Spent:       spent,
		Remaining:   remaining,
		PercentUsed: percentUsed,
		Status:      status,
		Categories:  categories,
		Tags:        tags,
	}, nil
}

func (s *Service) calculateSpending(budget *gen.Budget) (int64, error) {
	var catAmount int64
	if budget.AllCategories == 1 {
		total, err := s.q.SumAllCategoryExpenses(s.ctx(), gen.SumAllCategoryExpensesParams{
			PeriodStart: budget.PeriodStart,
			PeriodEnd:   budget.PeriodEnd,
		})
		if err != nil {
			return 0, err
		}
		catAmount = toInt64(total)
	} else {
		total, err := s.q.SumCategoryExpenses(s.ctx(), gen.SumCategoryExpensesParams{
			BudgetID:    budget.ID,
			PeriodStart: budget.PeriodStart,
			PeriodEnd:   budget.PeriodEnd,
		})
		if err != nil {
			return 0, err
		}
		catAmount = toInt64(total)
	}

	tagParams := gen.SumTagExpensesParams{
		BudgetID:    budget.ID,
		PeriodStart: budget.PeriodStart,
		PeriodEnd:   budget.PeriodEnd,
	}
	tagTotal, err := s.q.SumTagExpenses(s.ctx(), tagParams)
	if err != nil {
		return 0, err
	}
	tagAmount := toInt64(tagTotal)

	return catAmount + tagAmount, nil
}

func toInt64(v interface{}) int64 {
	if b, ok := v.(int64); ok {
		return b
	}
	return 0
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (s *Service) EditBudget(id int64, params EditBudgetParams) (*BudgetResult, error) {
	_, err := s.q.GetBudgetByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "budget", Name: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	var name sql.NullString
	if params.Name != "" {
		name = sql.NullString{String: params.Name, Valid: true}
	}

	var amountVal sql.NullInt64
	if params.Amount != nil {
		if *params.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		amountVal = sql.NullInt64{Int64: *params.Amount, Valid: true}
	}

	var notifyVal sql.NullInt64
	if params.NotifyPct != nil {
		if *params.NotifyPct < 1 || *params.NotifyPct > 100 {
			return nil, &ValidationError{Field: "notify", Message: "notification threshold must be between 1 and 100"}
		}
		notifyVal = sql.NullInt64{Int64: *params.NotifyPct, Valid: true}
	}

	updated, err := s.q.UpdateBudget(s.ctx(), gen.UpdateBudgetParams{
		ID:          id,
		Name:        name,
		Amount:      amountVal,
		NotifyAtPct: notifyVal,
	})
	if err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}

	for _, catName := range params.AddCategories {
		cat, err := s.ResolveCategory(catName)
		if err != nil {
			return nil, fmt.Errorf("add category '%s': %w", catName, err)
		}
		if err := s.q.AddBudgetCategory(s.ctx(), gen.AddBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget category: %w", err)
		}
	}

	for _, catName := range params.RemoveCategories {
		cat, err := s.ResolveCategory(catName)
		if err != nil {
			return nil, fmt.Errorf("remove category '%s': %w", catName, err)
		}
		if err := s.q.RemoveBudgetCategory(s.ctx(), gen.RemoveBudgetCategoryParams{
			BudgetID:   id,
			CategoryID: cat.ID,
		}); err != nil {
			return nil, fmt.Errorf("remove budget category: %w", err)
		}
	}

	for _, tagName := range params.AddTags {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("add tag '%s': %w", tagName, err)
		}
		if err := s.q.AddBudgetTag(s.ctx(), gen.AddBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("add budget tag: %w", err)
		}
	}

	for _, tagName := range params.RemoveTags {
		tag, err := s.ResolveTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("remove tag '%s': %w", tagName, err)
		}
		if err := s.q.RemoveBudgetTag(s.ctx(), gen.RemoveBudgetTagParams{
			BudgetID: id,
			TagID:    tag.ID,
		}); err != nil {
			return nil, fmt.Errorf("remove budget tag: %w", err)
		}
	}

	categories, _ := s.q.ListBudgetCategories(s.ctx(), updated.ID)
	tags, _ := s.q.ListBudgetTags(s.ctx(), updated.ID)

	return &BudgetResult{
		Budget:     updated,
		Categories: categories,
		Tags:       tags,
	}, nil
}

func (s *Service) RemoveBudget(id int64) error {
	_, err := s.q.GetBudgetByID(s.ctx(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{Entity: "budget", Name: fmt.Sprintf("%d", id)}
		}
		return err
	}
	return s.q.MarkBudgetInactive(s.ctx(), id)
}

func budgetName(b *gen.Budget) string {
	if b.Name.Valid {
		return b.Name.String
	}
	return fmt.Sprintf("%d", b.ID)
}
