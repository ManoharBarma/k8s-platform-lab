// Package service holds the business logic layer. Handlers call into Service;
// Service orchestrates domain rules and delegates persistence to a Repository.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/templates/service-template/internal/domain"
)

// Repository abstracts persistence for Resource aggregates. Implementations
// live in the repository package (in-memory and Postgres).
type Repository interface {
	Create(ctx context.Context, r domain.Resource) (domain.Resource, error)
	Get(ctx context.Context, id string) (domain.Resource, error)
	List(ctx context.Context) ([]domain.Resource, error)
}

// Service implements the template's use cases.
type Service struct {
	repo Repository
	log  *zap.Logger
}

// New builds a Service.
func New(repo Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create validates input and persists a new Resource.
func (s *Service) Create(ctx context.Context, name string) (domain.Resource, error) {
	if name == "" {
		return domain.Resource{}, apierror.ErrBadRequest("name is required")
	}
	r := domain.Resource{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	return s.repo.Create(ctx, r)
}

// Get returns a Resource by ID.
func (s *Service) Get(ctx context.Context, id string) (domain.Resource, error) {
	return s.repo.Get(ctx, id)
}

// List returns all Resources.
func (s *Service) List(ctx context.Context) ([]domain.Resource, error) {
	return s.repo.List(ctx)
}
