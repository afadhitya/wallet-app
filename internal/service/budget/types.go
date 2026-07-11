package budget

import "github.com/afadhitya/wallet-app/internal/gen"

type SetBudgetParams struct {
	Name       string
	Amount     int64
	Period     string
	From       string
	To         string
	NotifyPct  int64
	Categories []string
	Tags       []string
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
