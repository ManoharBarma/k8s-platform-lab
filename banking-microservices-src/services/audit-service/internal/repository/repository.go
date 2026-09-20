// Package repository provides append-only persistence for audit events.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/services/audit-service/internal/domain"
)

// Repository is an append-only store of audit events.
type Repository interface {
	Append(ctx context.Context, e domain.Event) error
	List(ctx context.Context, topic string, limit int) ([]domain.Event, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory audit log.
type InMemory struct {
	mu     sync.RWMutex
	events []domain.Event
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{} }

func (r *InMemory) Append(_ context.Context, e domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}

func (r *InMemory) List(_ context.Context, topic string, limit int) ([]domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Event
	for i := len(r.events) - 1; i >= 0; i-- {
		if topic == "" || r.events[i].Topic == topic {
			out = append(out, r.events[i])
			if len(out) >= limit {
				break
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists audit events under the audit_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Append(ctx context.Context, e domain.Event) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_events (id, topic, key, payload, created_at) VALUES ($1,$2,$3,$4,$5)`,
		e.ID, e.Topic, e.Key, []byte(e.Payload), e.CreatedAt)
	return err
}

func (r *Postgres) List(ctx context.Context, topic string, limit int) ([]domain.Event, error) {
	var rows interface {
		Next() bool
		Scan(...any) error
		Err() error
		Close()
	}
	var err error
	if topic == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, topic, key, payload, created_at FROM audit_events ORDER BY created_at DESC LIMIT $1`, limit)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, topic, key, payload, created_at FROM audit_events WHERE topic=$1 ORDER BY created_at DESC LIMIT $2`, topic, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Event
	for rows.Next() {
		var e domain.Event
		var payload []byte
		if err := rows.Scan(&e.ID, &e.Topic, &e.Key, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = payload
		out = append(out, e)
	}
	return out, rows.Err()
}
