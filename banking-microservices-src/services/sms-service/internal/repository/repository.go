// Package repository provides persistence for sms-service.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/services/sms-service/internal/domain"
)

// Repository stores SMS messages.
type Repository interface {
	Create(ctx context.Context, m domain.Message) error
	List(ctx context.Context, limit int) ([]domain.Message, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory message store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Message
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Message{}} }

func (r *InMemory) Create(_ context.Context, m domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.ID] = m
	return nil
}

func (r *InMemory) List(_ context.Context, limit int) ([]domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Message
	for _, m := range r.items {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// --- Postgres ---

// Postgres persists messages under the sms_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, to_number, body, status, created_at"

func (r *Postgres) Create(ctx context.Context, m domain.Message) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sms_messages (id, to_number, body, status, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		m.ID, m.To, m.Body, m.Status, m.CreatedAt)
	return err
}

func (r *Postgres) List(ctx context.Context, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM sms_messages ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.To, &m.Body, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
