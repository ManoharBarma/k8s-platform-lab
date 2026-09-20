// Package service implements audit-service: it records every consumed domain
// event into an append-only log and serves queries over it.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/audit-service/internal/domain"
	"banking-platform/services/audit-service/internal/repository"
)

// Service holds audit-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Handle is the Kafka consumer callback: it appends the event to the audit log.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	return s.repo.Append(ctx, domain.Event{
		ID:        uuid.NewString(),
		Topic:     m.Topic,
		Key:       m.Key,
		Payload:   m.Value,
		CreatedAt: time.Now().UTC(),
	})
}

// List returns recent audit events, optionally filtered by topic.
func (s *Service) List(ctx context.Context, topic string, limit int) ([]domain.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.List(ctx, topic, limit)
}
