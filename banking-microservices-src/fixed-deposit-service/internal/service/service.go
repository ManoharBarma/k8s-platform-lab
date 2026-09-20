// Package service implements fixed-deposit-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/fixed-deposit-service/internal/domain"
	"banking-platform/services/fixed-deposit-service/internal/repository"
)

// Service holds fixed-deposit-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create opens a fixed deposit for an owner.
func (s *Service) Create(ctx context.Context, ownerID, principalStr, currency string, annualRatePct float64, tenureMonths int) (domain.FixedDeposit, error) {
	principal, err := decimal.NewFromString(principalStr)
	if err != nil || principal.LessThanOrEqual(decimal.Zero) {
		return domain.FixedDeposit{}, apierror.ErrBadRequest("principal must be a positive amount")
	}
	if tenureMonths <= 0 {
		return domain.FixedDeposit{}, apierror.ErrBadRequest("tenure_months must be positive")
	}
	if currency == "" {
		currency = "USD"
	}
	now := time.Now().UTC()
	fd := domain.FixedDeposit{
		ID:             uuid.NewString(),
		OwnerID:        ownerID,
		Principal:      principal,
		Currency:       currency,
		AnnualRatePct:  annualRatePct,
		TenureMonths:   tenureMonths,
		MaturityAmount: domain.CalculateMaturity(principal, annualRatePct, tenureMonths),
		Status:         "ACTIVE",
		CreatedAt:      now,
		MaturesAt:      now.AddDate(0, tenureMonths, 0),
	}
	if err := s.repo.Create(ctx, fd); err != nil {
		return domain.FixedDeposit{}, err
	}
	return fd, nil
}

// GetOwned returns a fixed deposit, enforcing ownership.
func (s *Service) GetOwned(ctx context.Context, ownerID, id string) (domain.FixedDeposit, error) {
	fd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.FixedDeposit{}, err
	}
	if fd.OwnerID != ownerID {
		return domain.FixedDeposit{}, apierror.ErrForbidden("not your fixed deposit")
	}
	return fd, nil
}

// ListByOwner returns an owner's fixed deposits.
func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]domain.FixedDeposit, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Close marks an owner's fixed deposit as closed.
func (s *Service) Close(ctx context.Context, ownerID, id string) (domain.FixedDeposit, error) {
	fd, err := s.GetOwned(ctx, ownerID, id)
	if err != nil {
		return domain.FixedDeposit{}, err
	}
	fd.Status = "CLOSED"
	if err := s.repo.Update(ctx, fd); err != nil {
		return domain.FixedDeposit{}, err
	}
	return fd, nil
}
