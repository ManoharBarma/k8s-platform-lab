// Package service implements kyc-service use cases: submission and review.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/kyc-service/internal/domain"
	"banking-platform/services/kyc-service/internal/repository"
)

// Service holds kyc-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Submit records a new KYC submission in PENDING state.
func (s *Service) Submit(ctx context.Context, userID, docType, docID string) (domain.Verification, error) {
	if docType == "" || docID == "" {
		return domain.Verification{}, apierror.ErrBadRequest("document_type and document_id are required")
	}
	v := domain.Verification{
		ID:           uuid.NewString(),
		UserID:       userID,
		DocumentType: docType,
		DocumentID:   docID,
		Status:       domain.StatusPending,
		SubmittedAt:  time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return domain.Verification{}, err
	}
	return v, nil
}

// LatestForUser returns a user's most recent KYC verification.
func (s *Service) LatestForUser(ctx context.Context, userID string) (domain.Verification, error) {
	return s.repo.LatestByUser(ctx, userID)
}

// Pending lists submissions awaiting review.
func (s *Service) Pending(ctx context.Context) ([]domain.Verification, error) {
	return s.repo.ListByStatus(ctx, domain.StatusPending)
}

// Review approves or rejects a submission.
func (s *Service) Review(ctx context.Context, id string, approve bool, reason string) (domain.Verification, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Verification{}, err
	}
	if v.Status != domain.StatusPending {
		return domain.Verification{}, apierror.ErrConflict("verification already reviewed")
	}
	now := time.Now().UTC()
	v.ReviewedAt = &now
	v.Reason = reason
	if approve {
		v.Status = domain.StatusVerified
	} else {
		v.Status = domain.StatusRejected
	}
	if err := s.repo.UpdateReview(ctx, v); err != nil {
		return domain.Verification{}, err
	}
	return v, nil
}
