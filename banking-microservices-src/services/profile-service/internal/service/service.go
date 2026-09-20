// Package service implements profile-service use cases.
package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"banking-platform/services/profile-service/internal/domain"
	"banking-platform/services/profile-service/internal/repository"
)

// Service holds profile-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Get returns the profile for a user.
func (s *Service) Get(ctx context.Context, userID string) (domain.Profile, error) {
	return s.repo.Get(ctx, userID)
}

// Upsert creates or replaces a user's profile.
func (s *Service) Upsert(ctx context.Context, p domain.Profile) (domain.Profile, error) {
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Upsert(ctx, p); err != nil {
		return domain.Profile{}, err
	}
	return p, nil
}
