// Package service implements RBAC authorization: permission checks, role
// lookups, assignments and seeding of the default role model.
package service

import (
	"context"

	"go.uber.org/zap"

	"banking-platform/services/authz-service/internal/domain"
	"banking-platform/services/authz-service/internal/repository"
)

// Service holds authz-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Check reports whether subject holds permission. The wildcard permission "*"
// (held by admin) grants everything.
func (s *Service) Check(ctx context.Context, subject, permission string) (bool, error) {
	roles, err := s.repo.RolesForSubject(ctx, subject)
	if err != nil {
		return false, err
	}
	perms, err := s.repo.PermissionsForRoles(ctx, roles)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == domain.Wildcard || p == permission {
			return true, nil
		}
	}
	return false, nil
}

// GetRoles returns the roles assigned to a subject.
func (s *Service) GetRoles(ctx context.Context, subject string) ([]string, error) {
	return s.repo.RolesForSubject(ctx, subject)
}

// AssignRole grants a role to a subject.
func (s *Service) AssignRole(ctx context.Context, subject, role string) error {
	return s.repo.AssignRole(ctx, subject, role)
}

// ListRoles returns all defined roles with their permissions.
func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.repo.ListRoles(ctx)
}

// Seed installs the platform's default role/permission model. It is idempotent.
func (s *Service) Seed(ctx context.Context) error {
	permissions := map[string]string{
		"customer:read":     "Read customer records",
		"customer:write":    "Modify customer records",
		"account:read":      "Read accounts",
		"account:write":     "Open/modify accounts",
		"transaction:read":  "Read transactions",
		"transaction:write": "Create transactions",
		domain.Wildcard:     "All permissions (superuser)",
	}
	for name, desc := range permissions {
		if err := s.repo.EnsurePermission(ctx, name, desc); err != nil {
			return err
		}
	}

	roles := map[string]struct {
		desc  string
		perms []string
	}{
		"customer": {"End customer", []string{"customer:read", "account:read", "transaction:read"}},
		"teller":   {"Bank teller", []string{"customer:read", "customer:write", "account:read", "account:write", "transaction:read", "transaction:write"}},
		"admin":    {"Platform administrator", []string{domain.Wildcard}},
	}
	for name, def := range roles {
		if err := s.repo.EnsureRole(ctx, name, def.desc); err != nil {
			return err
		}
		for _, p := range def.perms {
			if err := s.repo.GrantPermission(ctx, name, p); err != nil {
				return err
			}
		}
	}
	s.log.Info("seeded default RBAC roles and permissions")
	return nil
}
