package repository

import (
	"context"
	"sort"
	"sync"

	"banking-platform/services/authz-service/internal/domain"
)

// InMemory is a concurrency-safe in-memory RBAC store.
type InMemory struct {
	mu          sync.RWMutex
	roles       map[string]string              // role -> description
	permissions map[string]string              // permission -> description
	rolePerms   map[string]map[string]struct{} // role -> permissions
	userRoles   map[string]map[string]struct{} // subject -> roles
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{
		roles:       map[string]string{},
		permissions: map[string]string{},
		rolePerms:   map[string]map[string]struct{}{},
		userRoles:   map[string]map[string]struct{}{},
	}
}

func (r *InMemory) EnsureRole(_ context.Context, name, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[name] = description
	if r.rolePerms[name] == nil {
		r.rolePerms[name] = map[string]struct{}{}
	}
	return nil
}

func (r *InMemory) EnsurePermission(_ context.Context, name, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[name] = description
	return nil
}

func (r *InMemory) GrantPermission(_ context.Context, role, permission string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.rolePerms[role] == nil {
		r.rolePerms[role] = map[string]struct{}{}
	}
	r.rolePerms[role][permission] = struct{}{}
	return nil
}

func (r *InMemory) AssignRole(_ context.Context, subject, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.userRoles[subject] == nil {
		r.userRoles[subject] = map[string]struct{}{}
	}
	r.userRoles[subject][role] = struct{}{}
	return nil
}

func (r *InMemory) RolesForSubject(_ context.Context, subject string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return keys(r.userRoles[subject]), nil
}

func (r *InMemory) PermissionsForRoles(_ context.Context, roles []string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	set := map[string]struct{}{}
	for _, role := range roles {
		for p := range r.rolePerms[role] {
			set[p] = struct{}{}
		}
	}
	return keys(set), nil
}

func (r *InMemory) ListRoles(_ context.Context) ([]domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Role, 0, len(r.roles))
	for name, desc := range r.roles {
		out = append(out, domain.Role{Name: name, Description: desc, Permissions: keys(r.rolePerms[name])})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
