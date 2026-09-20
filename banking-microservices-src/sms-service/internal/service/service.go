// Package service implements sms-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/sms-service/internal/domain"
	"banking-platform/services/sms-service/internal/repository"
)

// Service holds sms-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Send dispatches a mock SMS and stores it.
func (s *Service) Send(ctx context.Context, to, body string) (domain.Message, error) {
	if to == "" {
		return domain.Message{}, apierror.ErrBadRequest("to is required")
	}
	m := domain.Message{
		ID:        uuid.NewString(),
		To:        to,
		Body:      body,
		Status:    "SENT",
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return domain.Message{}, err
	}
	s.log.Info("sms sent", zap.String("id", m.ID), zap.String("to", m.To))
	return m, nil
}

// List returns recent SMS messages, newest first.
func (s *Service) List(ctx context.Context, limit int) ([]domain.Message, error) {
	return s.repo.List(ctx, limit)
}
