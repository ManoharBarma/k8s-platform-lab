// Package service implements fraud-service: it consumes domain events, applies
// simple fraud rules, and records an alert when a rule is tripped.
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/fraud-service/internal/domain"
	"banking-platform/services/fraud-service/internal/repository"
)

// Service holds fraud-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// event is the union of the transaction and payment event shapes; fields absent
// in a given event stay at their zero value.
type event struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	PayerAccountID string `json:"payer_account_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
}

// Handle is the Kafka consumer callback: it applies fraud rules and records an
// alert when one is tripped.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	var ev event
	if err := json.Unmarshal(m.Value, &ev); err != nil {
		s.log.Warn("skipping unparseable event", zap.String("topic", m.Topic), zap.Error(err))
		return nil
	}

	accountID := ev.FromAccountID
	if accountID == "" {
		accountID = ev.PayerAccountID
	}

	amount, err := decimal.NewFromString(ev.Amount)
	if err != nil {
		s.log.Warn("skipping event with invalid amount", zap.String("amount", ev.Amount), zap.Error(err))
		return nil
	}

	var score int
	var reason string
	switch {
	case amount.GreaterThanOrEqual(decimal.NewFromInt(10000)):
		score, reason = 90, "large amount"
	case amount.GreaterThanOrEqual(decimal.NewFromInt(5000)) && ev.Type == "WITHDRAWAL":
		score, reason = 70, "large withdrawal"
	default:
		return nil
	}

	alert := domain.Alert{
		ID:            uuid.NewString(),
		TransactionID: ev.ID,
		AccountID:     accountID,
		Amount:        amount,
		Currency:      ev.Currency,
		Reason:        reason,
		Score:         score,
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, alert); err != nil {
		return err
	}
	s.log.Info("fraud alert raised",
		zap.String("transaction_id", alert.TransactionID),
		zap.String("account_id", alert.AccountID),
		zap.Int("score", alert.Score),
		zap.String("reason", alert.Reason))
	return nil
}

// List returns recent fraud alerts, newest first.
func (s *Service) List(ctx context.Context, limit int) ([]domain.Alert, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.List(ctx, limit)
}
