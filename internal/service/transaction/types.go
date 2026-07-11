package transaction

import "github.com/afadhitya/wallet-app/internal/gen"

type CreateExpenseParams struct {
	Amount      int64
	Description string
	Category    string
	Account     string
	Tags        []string
	Date        string
	Notes       string
}

type CreateIncomeParams struct {
	Amount      int64
	Description string
	Category    string
	Account     string
	Tags        []string
	Date        string
	Notes       string
}

type TransactionResult struct {
	Transaction *gen.Transaction
	Tags        []*gen.Tag
}

type CreateTransferParams struct {
	Amount      int64
	FromAccount string
	ToAccount   string
	Date        string
	Description string
	Notes       string
}

type TransferResult struct {
	Transaction *gen.Transaction
	Warning     string
}

type ListTransactionsParams struct {
	AccountName  string
	CategoryName string
	TagName      string
	Type         string
	Month        string
	DateFrom     string
	DateTo       string
	Limit        int
	Offset       int
}

type ListTransactionsResult struct {
	Transactions []*gen.Transaction
	Total        int64
	BaseTotal    int64
	Currency     string
}

type EditTransactionParams struct {
	Amount         *int64
	CategoryName   string
	AccountName    string
	Date           string
	Description    string
	Notes          string
	AddTagNames    []string
	RemoveTagNames []string
}

type AdjustBalanceParams struct {
	Account     string
	Target      int64
	Description string
	Notes       string
}

type AdjustBalanceResult struct {
	Account     *gen.Account
	OldBalance  int64
	NewBalance  int64
	Difference  int64
	Transaction *gen.Transaction
}
