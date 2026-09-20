// Package service implements email-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/email-service/internal/domain"
	"banking-platform/services/email-service/internal/repository"
)

// Service holds email-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Send stores a mock email and marks it as sent.
func (s *Service) Send(ctx context.Context, to, subject, body string) (domain.Message, error) {
	if to == "" {
		return domain.Message{}, apierror.ErrBadRequest("to is required")
	}
	m := domain.Message{
		ID:        uuid.NewString(),
		To:        to,
		Subject:   subject,
		Body:      body,
		Status:    "SENT",
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return domain.Message{}, err
	}
	s.log.Info("email sent", zap.String("id", m.ID), zap.String("to", m.To))
	return m, nil
}

// List returns the most recent messages.
func (s *Service) List(ctx context.Context, limit int) ([]domain.Message, error) {
	return s.repo.List(ctx, limit)
}
