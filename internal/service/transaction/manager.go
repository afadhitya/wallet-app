package transaction

import "github.com/afadhitya/wallet-app/internal/gen"

type Manager struct {
	q gen.Querier
}

func NewManager(q gen.Querier) *Manager {
	return &Manager{q: q}
}
