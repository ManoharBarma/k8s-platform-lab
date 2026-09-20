// Package repository provides persistence for auth-service.
package repository

import (
	"context"

	"banking-platform/services/auth-service/internal/domain"
)

// Repository stores login identities.
type Repository interface {
	Create(ctx context.Context, u domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
}
