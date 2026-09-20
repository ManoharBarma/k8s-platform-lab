// Package repository provides persistence implementations for the service.
package repository

import (
	"context"
	"sync"

	"banking-platform/pkg/apierror"
	"banking-platform/templates/service-template/internal/domain"
)

// InMemory is a concurrency-safe, in-memory Repository used when Postgres is
// disabled (e.g. quick local runs and tests).
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Resource
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{items: make(map[string]domain.Resource)}
}

func (r *InMemory) Create(_ context.Context, res domain.Resource) (domain.Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[res.ID] = res
	return res, nil
}

func (r *InMemory) Get(_ context.Context, id string) (domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.items[id]
	if !ok {
		return domain.Resource{}, apierror.ErrNotFound("resource not found")
	}
	return res, nil
}

func (r *InMemory) List(_ context.Context) ([]domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Resource, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, v)
	}
	return out, nil
}
