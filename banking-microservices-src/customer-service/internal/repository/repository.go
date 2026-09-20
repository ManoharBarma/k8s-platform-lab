// Package repository provides persistence for customer-service.
package repository

import (
	"context"

	"banking-platform/services/customer-service/internal/domain"
)

// Repository stores customer master records.
type Repository interface {
	Create(ctx context.Context, c domain.Customer) error
	GetByID(ctx context.Context, id string) (domain.Customer, error)
	GetByUserID(ctx context.Context, userID string) (domain.Customer, error)
	Update(ctx context.Context, c domain.Customer) error
	List(ctx context.Context, limit, offset int) ([]domain.Customer, error)
}
