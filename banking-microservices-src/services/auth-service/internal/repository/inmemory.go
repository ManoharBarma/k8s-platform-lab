package repository

import (
	"context"
	"strings"
	"sync"

	"banking-platform/pkg/apierror"
	"banking-platform/services/auth-service/internal/domain"
)

// InMemory is a concurrency-safe in-memory user store (used when Postgres is
// disabled).
type InMemory struct {
	mu       sync.RWMutex
	byID     map[string]domain.User
	byEmail  map[string]string // email -> id
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{byID: map[string]domain.User{}, byEmail: map[string]string{}}
}

func (r *InMemory) Create(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToLower(u.Email)
	if _, ok := r.byEmail[key]; ok {
		return apierror.ErrConflict("email already registered")
	}
	r.byID[u.ID] = u
	r.byEmail[key] = u.ID
	return nil
}

func (r *InMemory) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return domain.User{}, apierror.ErrNotFound("user not found")
	}
	return r.byID[id], nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.User{}, apierror.ErrNotFound("user not found")
	}
	return u, nil
}
