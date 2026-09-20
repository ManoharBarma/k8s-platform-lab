// Package service implements admin-service use cases.
package service

import (
	"context"

	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/admin-service/internal/domain"
	"banking-platform/services/admin-service/internal/repository"
)

// Service holds admin-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Seed installs the platform's default settings. It is idempotent.
func (s *Service) Seed(ctx context.Context) error {
	defaults := map[string]string{
		"maintenance_mode":   "false",
		"signup_enabled":     "true",
		"max_daily_transfer": "50000",
	}
	for key, value := range defaults {
		if err := s.repo.Upsert(ctx, key, value); err != nil {
			return err
		}
	}
	s.log.Info("seeded admin settings")
	return nil
}

// List returns all settings ordered by key.
func (s *Service) List(ctx context.Context) ([]domain.Setting, error) {
	return s.repo.List(ctx)
}

// Get returns a single setting by key.
func (s *Service) Get(ctx context.Context, key string) (domain.Setting, error) {
	return s.repo.Get(ctx, key)
}

// Set upserts a setting and returns the stored value.
func (s *Service) Set(ctx context.Context, key, value string) (domain.Setting, error) {
	if key == "" {
		return domain.Setting{}, apierror.ErrBadRequest("key is required")
	}
	if err := s.repo.Upsert(ctx, key, value); err != nil {
		return domain.Setting{}, err
	}
	return s.repo.Get(ctx, key)
}
