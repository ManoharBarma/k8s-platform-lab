// Package service implements beneficiary-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/beneficiary-service/internal/domain"
	"banking-platform/services/beneficiary-service/internal/repository"
)

// Service holds beneficiary-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Add creates a beneficiary for an owner.
func (s *Service) Add(ctx context.Context, b domain.Beneficiary) (domain.Beneficiary, error) {
	if b.Name == "" || b.AccountNumber == "" {
		return domain.Beneficiary{}, apierror.ErrBadRequest("name and account_number are required")
	}
	b.ID = uuid.NewString()
	b.CreatedAt = time.Now().UTC()
	if err := s.repo.Create(ctx, b); err != nil {
		return domain.Beneficiary{}, err
	}
	return b, nil
}

// List returns an owner's beneficiaries.
func (s *Service) List(ctx context.Context, ownerID string) ([]domain.Beneficiary, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Get returns a beneficiary, enforcing ownership.
func (s *Service) Get(ctx context.Context, ownerID, id string) (domain.Beneficiary, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Beneficiary{}, err
	}
	if b.OwnerID != ownerID {
		return domain.Beneficiary{}, apierror.ErrForbidden("not your beneficiary")
	}
	return b, nil
}

// Delete removes an owner's beneficiary.
func (s *Service) Delete(ctx context.Context, ownerID, id string) error {
	if _, err := s.Get(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
