// Package repository provides persistence for the RBAC model.
package repository

import (
	"context"

	"banking-platform/services/authz-service/internal/domain"
)

// Repository stores roles, permissions and their assignments. All mutating
// methods are idempotent so they can be used for seeding.
type Repository interface {
	EnsureRole(ctx context.Context, name, description string) error
	EnsurePermission(ctx context.Context, name, description string) error
	GrantPermission(ctx context.Context, role, permission string) error
	AssignRole(ctx context.Context, subject, role string) error
	RolesForSubject(ctx context.Context, subject string) ([]string, error)
	PermissionsForRoles(ctx context.Context, roles []string) ([]string, error)
	ListRoles(ctx context.Context) ([]domain.Role, error)
}
