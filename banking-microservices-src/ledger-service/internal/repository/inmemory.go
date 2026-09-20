package repository

import (
	"context"
	"sync"

	"github.com/shopspring/decimal"

	"banking-platform/pkg/apierror"
	"banking-platform/services/ledger-service/internal/domain"
)

// InMemory is a concurrency-safe in-memory ledger (used when Postgres is
// disabled). It preserves the same invariants as the Postgres implementation.
type InMemory struct {
	mu          sync.Mutex
	accounts    map[string]*domain.Account
	appliedKeys map[string]struct{} // idempotency keys already applied
}

// NewInMemory builds an empty in-memory ledger.
func NewInMemory() *InMemory {
	return &InMemory{
		accounts:    map[string]*domain.Account{},
		appliedKeys: map[string]struct{}{},
	}
}

func (r *InMemory) OpenAccount(_ context.Context, accountID, currency string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accounts[accountID]; ok {
		return nil // idempotent
	}
	r.accounts[accountID] = &domain.Account{AccountID: accountID, Currency: currency, Balance: decimal.Zero}
	return nil
}

func (r *InMemory) Post(_ context.Context, t domain.Transaction) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.appliedKeys[t.IdempotencyKey]; ok {
		return true, nil
	}

	// Validate all accounts exist before mutating any balance.
	for _, e := range t.Entries {
		if _, ok := r.accounts[e.AccountID]; !ok {
			return false, apierror.ErrBadRequest("ledger account not found: " + e.AccountID)
		}
	}
	for _, e := range t.Entries {
		acc := r.accounts[e.AccountID]
		if e.Direction == domain.Credit {
			acc.Balance = acc.Balance.Add(e.Amount)
		} else {
			acc.Balance = acc.Balance.Sub(e.Amount)
		}
	}
	r.appliedKeys[t.IdempotencyKey] = struct{}{}
	return false, nil
}

func (r *InMemory) Balance(_ context.Context, accountID string) (decimal.Decimal, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acc, ok := r.accounts[accountID]
	if !ok {
		return decimal.Zero, "", apierror.ErrNotFound("ledger account not found")
	}
	return acc.Balance, acc.Currency, nil
}
