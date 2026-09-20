// Package service implements recurring-deposit-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/recurring-deposit-service/internal/domain"
	"banking-platform/services/recurring-deposit-service/internal/repository"
)

// Service holds recurring-deposit-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create opens a recurring deposit for an owner.
func (s *Service) Create(ctx context.Context, ownerID, monthlyStr, currency string, annualRatePct float64, tenureMonths int) (domain.RecurringDeposit, error) {
	monthly, err := decimal.NewFromString(monthlyStr)
	if err != nil {
		return domain.RecurringDeposit{}, apierror.ErrBadRequest("invalid monthly_amount")
	}
	if monthly.LessThanOrEqual(decimal.Zero) {
		return domain.RecurringDeposit{}, apierror.ErrBadRequest("monthly_amount must be positive")
	}
	if tenureMonths <= 0 {
		return domain.RecurringDeposit{}, apierror.ErrBadRequest("tenure_months must be positive")
	}
	if currency == "" {
		currency = "USD"
	}
	d := domain.RecurringDeposit{
		ID:               uuid.NewString(),
		OwnerID:          ownerID,
		MonthlyAmount:    monthly,
		Currency:         currency,
		AnnualRatePct:    annualRatePct,
		TenureMonths:     tenureMonths,
		InstallmentsPaid: 0,
		MaturityAmount:   domain.CalculateMaturity(monthly, annualRatePct, tenureMonths),
		Status:           "ACTIVE",
		CreatedAt:        time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return domain.RecurringDeposit{}, err
	}
	return d, nil
}

// Deposit records a single installment, enforcing ownership.
func (s *Service) Deposit(ctx context.Context, ownerID, id string) (domain.RecurringDeposit, error) {
	d, err := s.GetOwned(ctx, ownerID, id)
	if err != nil {
		return domain.RecurringDeposit{}, err
	}
	if d.InstallmentsPaid >= d.TenureMonths {
		return domain.RecurringDeposit{}, apierror.ErrConflict("all installments paid")
	}
	d.InstallmentsPaid++
	if d.InstallmentsPaid == d.TenureMonths {
		d.Status = "MATURED"
	}
	if err := s.repo.Update(ctx, d); err != nil {
		return domain.RecurringDeposit{}, err
	}
	return d, nil
}

// GetOwned returns a recurring deposit, enforcing ownership.
func (s *Service) GetOwned(ctx context.Context, ownerID, id string) (domain.RecurringDeposit, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.RecurringDeposit{}, err
	}
	if d.OwnerID != ownerID {
		return domain.RecurringDeposit{}, apierror.ErrForbidden("not your recurring deposit")
	}
	return d, nil
}

// ListByOwner returns an owner's recurring deposits.
func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]domain.RecurringDeposit, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}
