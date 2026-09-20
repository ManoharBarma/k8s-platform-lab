// Package service implements support-service use cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/support-service/internal/domain"
	"banking-platform/services/support-service/internal/repository"
)

// Service holds support-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Open creates a ticket for an owner.
func (s *Service) Open(ctx context.Context, ownerID, subject, body string) (domain.Ticket, error) {
	if subject == "" || body == "" {
		return domain.Ticket{}, apierror.ErrBadRequest("subject and body are required")
	}
	now := time.Now().UTC()
	t := domain.Ticket{
		ID:        uuid.NewString(),
		OwnerID:   ownerID,
		Subject:   subject,
		Body:      body,
		Status:    "OPEN",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	return t, nil
}

// GetOwned returns a ticket, enforcing ownership unless the caller is staff.
func (s *Service) GetOwned(ctx context.Context, ownerID, id string, isStaff bool) (domain.Ticket, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !isStaff && t.OwnerID != ownerID {
		return domain.Ticket{}, apierror.ErrForbidden("not your ticket")
	}
	return t, nil
}

// ListMine returns an owner's tickets.
func (s *Service) ListMine(ctx context.Context, ownerID string) ([]domain.Ticket, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// ListAll returns all tickets (staff).
func (s *Service) ListAll(ctx context.Context, limit int) ([]domain.Ticket, error) {
	return s.repo.ListAll(ctx, limit)
}

// UpdateStatus changes a ticket's status (staff action).
func (s *Service) UpdateStatus(ctx context.Context, id, status string) (domain.Ticket, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	return t, nil
}
