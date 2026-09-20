package repository

import (
	"context"
	"sort"
	"sync"

	"banking-platform/pkg/apierror"
	"banking-platform/services/customer-service/internal/domain"
)

// InMemory is a concurrency-safe in-memory customer store.
type InMemory struct {
	mu       sync.RWMutex
	byID     map[string]domain.Customer
	byUserID map[string]string
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{byID: map[string]domain.Customer{}, byUserID: map[string]string{}}
}

func (r *InMemory) Create(_ context.Context, c domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byUserID[c.UserID]; ok {
		return apierror.ErrConflict("customer already exists for user")
	}
	r.byID[c.ID] = c
	r.byUserID[c.UserID] = c.ID
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	if !ok {
		return domain.Customer{}, apierror.ErrNotFound("customer not found")
	}
	return c, nil
}

func (r *InMemory) GetByUserID(_ context.Context, userID string) (domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byUserID[userID]
	if !ok {
		return domain.Customer{}, apierror.ErrNotFound("customer not found")
	}
	return r.byID[id], nil
}

func (r *InMemory) Update(_ context.Context, c domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[c.ID]; !ok {
		return apierror.ErrNotFound("customer not found")
	}
	r.byID[c.ID] = c
	return nil
}

func (r *InMemory) List(_ context.Context, limit, offset int) ([]domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]domain.Customer, 0, len(r.byID))
	for _, c := range r.byID {
		all = append(all, c)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	if offset > len(all) {
		return []domain.Customer{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}
