// Package service implements analytics-service: it consumes transaction events
// from Kafka and maintains per-transaction-type aggregates, serving a summary.
package service

import (
	"context"
	"encoding/json"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/analytics-service/internal/domain"
	"banking-platform/services/analytics-service/internal/repository"
)

// Service holds analytics-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// event is the transaction event payload consumed from Kafka.
type event struct {
	Type     string          `json:"type"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// Handle is the Kafka consumer callback: it parses a transaction event and
// increments the aggregate for its type.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	var e event
	if err := json.Unmarshal(m.Value, &e); err != nil {
		s.log.Warn("skipping malformed event", zap.Error(err))
		return nil
	}
	if e.Type == "" {
		return nil
	}
	return s.repo.Add(ctx, e.Type, e.Amount)
}

// Summary returns the per-transaction-type aggregates ordered by key.
func (s *Service) Summary(ctx context.Context) ([]domain.Metric, error) {
	return s.repo.List(ctx)
}
