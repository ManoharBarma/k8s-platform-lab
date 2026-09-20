// Package service implements customer-service use cases.
package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/customer-service/internal/domain"
	"banking-platform/services/customer-service/internal/repository"
)

// Service holds customer-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// CreateInput is the data required to onboard a customer.
type CreateInput struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
	Phone     string
}

// Create onboards a new customer for the authenticated user.
func (s *Service) Create(ctx context.Context, in CreateInput) (domain.Customer, error) {
	if in.UserID == "" {
		return domain.Customer{}, apierror.ErrBadRequest("user id is required")
	}
	if strings.TrimSpace(in.FirstName) == "" {
		return domain.Customer{}, apierror.ErrBadRequest("first name is required")
	}
	now := time.Now().UTC()
	c := domain.Customer{
		ID:        uuid.NewString(),
		UserID:    in.UserID,
		Email:     strings.ToLower(in.Email),
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     in.Phone,
		Status:    domain.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return domain.Customer{}, err
	}
	return c, nil
}

// GetByUserID returns the customer linked to a login identity.
func (s *Service) GetByUserID(ctx context.Context, userID string) (domain.Customer, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetByID returns a customer by ID (admin/teller use).
func (s *Service) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a page of customers.
func (s *Service) List(ctx context.Context, limit, offset int) ([]domain.Customer, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

// UpdateInput is the mutable subset of a customer.
type UpdateInput struct {
	FirstName string
	LastName  string
	Phone     string
}

// UpdateByUserID updates the authenticated user's customer record.
func (s *Service) UpdateByUserID(ctx context.Context, userID string, in UpdateInput) (domain.Customer, error) {
	c, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Customer{}, err
	}
	if in.FirstName != "" {
		c.FirstName = in.FirstName
	}
	if in.LastName != "" {
		c.LastName = in.LastName
	}
	if in.Phone != "" {
		c.Phone = in.Phone
	}
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, c); err != nil {
		return domain.Customer{}, err
	}
	return c, nil
}
