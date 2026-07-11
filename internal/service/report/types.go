package report

type ReportParams struct {
	Month       string
	DateFrom    string
	DateTo      string
	AccountName string
	By          string
	Export      string
	OutputPath  string
}

type ReportFilters struct {
	DateFrom  string
	DateTo    string
	Period    string
	AccountID *int64
}

type ReportCategoryRow struct {
	CategoryID         int64   `json:"category_id"`
	CategoryName       string  `json:"category_name"`
	ParentCategoryName string  `json:"parent_category_name,omitempty"`
	Amount             int64   `json:"amount"`
	Percent            float64 `json:"percent"`
	TransactionCount   int64   `json:"transaction_count"`
}

type ReportAccountRow struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Currency    string `json:"currency"`
	Income      int64  `json:"income"`
	Expense     int64  `json:"expense"`
	Net         int64  `json:"net"`
}

type ReportTagRow struct {
	TagID            int64   `json:"tag_id"`
	TagName          string  `json:"tag_name"`
	Amount           int64   `json:"amount"`
	Percent          float64 `json:"percent"`
	TransactionCount int64   `json:"transaction_count"`
}

type ReportExportRow struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	BaseAmount  int64  `json:"base_amount"`
	Category    string `json:"category"`
	Account     string `json:"account"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

type ReportResult struct {
	BaseCurrency      string              `json:"base_currency"`
	Period            string              `json:"period"`
	IncomeTotal       int64               `json:"income_total"`
	ExpenseTotal      int64               `json:"expense_total"`
	Net               int64               `json:"net"`
	TransferTotal     int64               `json:"transfer_total"`
	TransactionCount  int64               `json:"transaction_count"`
	IncomeCategories  []ReportCategoryRow `json:"income_categories,omitempty"`
	ExpenseCategories []ReportCategoryRow `json:"expense_categories,omitempty"`
	ByCategory        []ReportCategoryRow `json:"by_category,omitempty"`
	ByAccount         []ReportAccountRow  `json:"by_account,omitempty"`
	ByTag             []ReportTagRow      `json:"by_tag,omitempty"`
	ExportRows        []ReportExportRow   `json:"export_rows,omitempty"`
}
