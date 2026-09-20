// Package service implements the ledger use cases and enforces double-entry
// invariants before anything is persisted.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/ledger-service/internal/domain"
	"banking-platform/services/ledger-service/internal/repository"
)

// Service holds ledger dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// OpenAccount opens a ledger account.
func (s *Service) OpenAccount(ctx context.Context, accountID, currency string) error {
	if accountID == "" || currency == "" {
		return apierror.ErrBadRequest("account_id and currency are required")
	}
	return s.repo.OpenAccount(ctx, accountID, currency)
}

// Post validates and applies a balanced transaction. It returns the transaction
// ID and whether the posting was a no-op duplicate.
func (s *Service) Post(ctx context.Context, t domain.Transaction) (string, bool, error) {
	if t.IdempotencyKey == "" {
		return "", false, apierror.ErrBadRequest("idempotency_key is required")
	}
	if t.Currency == "" {
		return "", false, apierror.ErrBadRequest("currency is required")
	}
	if len(t.Entries) < 2 {
		return "", false, apierror.ErrBadRequest("a transaction needs at least two entries")
	}

	debits, credits := decimal.Zero, decimal.Zero
	for _, e := range t.Entries {
		if !e.Direction.Valid() {
			return "", false, apierror.ErrBadRequest("invalid entry direction")
		}
		if e.Amount.LessThanOrEqual(decimal.Zero) {
			return "", false, apierror.ErrBadRequest("entry amount must be positive")
		}
		if e.Direction == domain.Debit {
			debits = debits.Add(e.Amount)
		} else {
			credits = credits.Add(e.Amount)
		}
	}
	if !debits.Equal(credits) {
		return "", false, apierror.ErrBadRequest("unbalanced transaction: debits must equal credits")
	}

	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	dup, err := s.repo.Post(ctx, t)
	if err != nil {
		return "", false, err
	}
	return t.ID, dup, nil
}

// Balance returns an account's balance.
func (s *Service) Balance(ctx context.Context, accountID string) (decimal.Decimal, string, error) {
	return s.repo.Balance(ctx, accountID)
}
