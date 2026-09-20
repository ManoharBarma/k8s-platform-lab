// Package service implements emi-service use cases: generating an amortization
// schedule for a loan and marking installments paid.
package service

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/emi-service/internal/domain"
	"banking-platform/services/emi-service/internal/repository"
)

// Service holds emi-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// CreatePlanInput describes a loan to amortize.
type CreatePlanInput struct {
	LoanID       string
	Principal    string
	AnnualRate   float64
	TenureMonths int
}

// CreatePlan generates and stores an amortization schedule for a loan.
func (s *Service) CreatePlan(ctx context.Context, in CreatePlanInput) ([]domain.Installment, error) {
	principal, err := decimal.NewFromString(in.Principal)
	if err != nil || principal.LessThanOrEqual(decimal.Zero) {
		return nil, apierror.ErrBadRequest("principal must be a positive number")
	}
	if in.TenureMonths <= 0 {
		return nil, apierror.ErrBadRequest("tenure_months must be positive")
	}
	if exists, _ := s.repo.HasSchedule(ctx, in.LoanID); exists {
		return nil, apierror.ErrConflict("schedule already exists for this loan")
	}
	schedule := domain.GenerateSchedule(in.LoanID, principal, in.AnnualRate, in.TenureMonths, time.Now().UTC())
	if err := s.repo.SaveSchedule(ctx, in.LoanID, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

// Schedule returns a loan's installments.
func (s *Service) Schedule(ctx context.Context, loanID string) ([]domain.Installment, error) {
	return s.repo.ListByLoan(ctx, loanID)
}

// PayNext marks the earliest unpaid installment as paid.
func (s *Service) PayNext(ctx context.Context, loanID string) (domain.Installment, error) {
	installments, err := s.repo.ListByLoan(ctx, loanID)
	if err != nil {
		return domain.Installment{}, err
	}
	for _, ins := range installments {
		if !ins.Paid {
			if err := s.repo.MarkPaid(ctx, ins.ID); err != nil {
				return domain.Installment{}, err
			}
			ins.Paid = true
			return ins, nil
		}
	}
	return domain.Installment{}, apierror.ErrConflict("no unpaid installments remaining")
}
