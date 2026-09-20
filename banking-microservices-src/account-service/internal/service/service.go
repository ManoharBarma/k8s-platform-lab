// Package service implements account-service use cases.
package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/account-service/internal/domain"
	"banking-platform/services/account-service/internal/repository"
)

// LedgerClient is the subset of the ledger the account service depends on.
type LedgerClient interface {
	OpenAccount(ctx context.Context, accountID, currency string) error
	Balance(ctx context.Context, accountID string) (balance, currency string, err error)
}

// Service holds account-service dependencies.
type Service struct {
	repo   repository.Repository
	ledger LedgerClient
	log    *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, ledger LedgerClient, log *zap.Logger) *Service {
	return &Service{repo: repo, ledger: ledger, log: log}
}

// Open creates a bank account and its backing ledger account.
func (s *Service) Open(ctx context.Context, customerID string, accType domain.Type, currency string) (domain.Account, error) {
	if accType != domain.TypeSavings && accType != domain.TypeCurrent {
		return domain.Account{}, apierror.ErrBadRequest("type must be SAVINGS or CURRENT")
	}
	if currency == "" {
		return domain.Account{}, apierror.ErrBadRequest("currency is required")
	}
	acc := domain.Account{
		ID:            uuid.NewString(),
		AccountNumber: generateAccountNumber(),
		CustomerID:    customerID,
		Type:          accType,
		Currency:      currency,
		Status:        domain.StatusActive,
		CreatedAt:     time.Now().UTC(),
	}
	// Open the ledger account first; without it the account cannot hold money.
	if err := s.ledger.OpenAccount(ctx, acc.ID, currency); err != nil {
		return domain.Account{}, apierror.New(502, "ledger_unavailable", "failed to open ledger account")
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		return domain.Account{}, err
	}
	return acc, nil
}

// Get returns an account by ID.
func (s *Service) Get(ctx context.Context, id string) (domain.Account, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByCustomer returns a customer's accounts.
func (s *Service) ListByCustomer(ctx context.Context, customerID string) ([]domain.Account, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

// Balance returns an account's balance from the ledger.
func (s *Service) Balance(ctx context.Context, id string) (string, string, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return "", "", err
	}
	bal, currency, err := s.ledger.Balance(ctx, id)
	if err != nil {
		return "", "", apierror.New(502, "ledger_unavailable", "failed to fetch balance")
	}
	return bal, currency, nil
}

// generateAccountNumber returns a random 12-digit account number.
func generateAccountNumber() string {
	const digits = 12
	buf := make([]byte, digits)
	for i := range buf {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		buf[i] = byte('0' + n.Int64())
	}
	return string(buf)
}
