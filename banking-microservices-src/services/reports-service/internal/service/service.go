// Package service implements reports-service: it consumes transaction events
// from Kafka and maintains per-day, per-currency aggregates.
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/reports-service/internal/domain"
	"banking-platform/services/reports-service/internal/repository"
)

// Service holds reports-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// transactionEvent is the subset of the transaction payload we aggregate on.
type transactionEvent struct {
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
}

// Handle is the Kafka consumer callback: it parses a transaction event and
// folds it into the daily aggregate for its (date, currency) bucket.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	var ev transactionEvent
	if err := json.Unmarshal(m.Value, &ev); err != nil {
		s.log.Warn("skipping unparseable event", zap.Error(err))
		return nil
	}
	if ev.CreatedAt.IsZero() {
		s.log.Warn("skipping event with missing created_at")
		return nil
	}
	date := ev.CreatedAt.UTC().Format("2006-01-02")
	return s.repo.Add(ctx, date, ev.Currency, ev.Amount)
}

// DailyList returns daily aggregates, optionally filtered by date.
func (s *Service) DailyList(ctx context.Context, date string) ([]domain.DailyTotal, error) {
	return s.repo.List(ctx, date)
}
