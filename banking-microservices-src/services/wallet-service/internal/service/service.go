// Package service implements wallet-service use cases: top-up and spend, both
// idempotent, with a transactional outbox event per movement.
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/outbox"
	"banking-platform/services/wallet-service/internal/domain"
	"banking-platform/services/wallet-service/internal/repository"
)

// EventTopic is the Kafka topic wallet events are published to.
const EventTopic = "banking.wallet"

// Service holds wallet-service dependencies.
type Service struct {
	repo       repository.Repository
	source     string
	emitEvents bool
	log        *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, source string, emitEvents bool, log *zap.Logger) *Service {
	return &Service{repo: repo, source: source, emitEvents: emitEvents, log: log}
}

// TopUp adds funds to a user's wallet (creating it if needed).
func (s *Service) TopUp(ctx context.Context, userID, currency, amountStr, idemKey string) (domain.Wallet, error) {
	amount, err := parsePositive(amountStr)
	if err != nil {
		return domain.Wallet{}, err
	}
	if currency == "" {
		currency = "USD"
	}
	if _, err := s.repo.Ensure(ctx, userID, currency); err != nil {
		return domain.Wallet{}, err
	}
	if _, ok := s.replay(ctx, idemKey); ok {
		return s.repo.Get(ctx, userID)
	}
	return s.apply(ctx, userID, domain.TxnTopUp, amount, amount, idemKey)
}

// Spend deducts funds from a user's wallet (rejecting overdrafts).
func (s *Service) Spend(ctx context.Context, userID, amountStr, idemKey string) (domain.Wallet, error) {
	amount, err := parsePositive(amountStr)
	if err != nil {
		return domain.Wallet{}, err
	}
	if _, ok := s.replay(ctx, idemKey); ok {
		return s.repo.Get(ctx, userID)
	}
	return s.apply(ctx, userID, domain.TxnSpend, amount, amount.Neg(), idemKey)
}

// Balance returns a user's wallet.
func (s *Service) Balance(ctx context.Context, userID string) (domain.Wallet, error) {
	return s.repo.Get(ctx, userID)
}

// History returns a user's wallet transactions.
func (s *Service) History(ctx context.Context, userID string) ([]domain.Transaction, error) {
	return s.repo.History(ctx, userID)
}

func (s *Service) apply(ctx context.Context, userID string, typ domain.TxnType, amount, delta decimal.Decimal, idemKey string) (domain.Wallet, error) {
	t := domain.Transaction{
		ID: uuid.NewString(), UserID: userID, Type: typ, Amount: amount,
		IdempotencyKey: idemKey, CreatedAt: time.Now().UTC(),
	}
	var event *outbox.Event
	if s.emitEvents {
		payload, _ := json.Marshal(t)
		event = &outbox.Event{
			ID: uuid.NewString(), Source: s.source, Topic: EventTopic, Key: userID, Payload: payload,
		}
	}
	return s.repo.Apply(ctx, t, delta, event)
}

func (s *Service) replay(ctx context.Context, idemKey string) (domain.Transaction, bool) {
	if idemKey == "" {
		return domain.Transaction{}, false
	}
	t, err := s.repo.TxnByKey(ctx, idemKey)
	if err != nil {
		return domain.Transaction{}, false
	}
	return t, true
}

func parsePositive(s string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, apierror.ErrBadRequest("invalid amount")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, apierror.ErrBadRequest("amount must be positive")
	}
	return amount, nil
}
