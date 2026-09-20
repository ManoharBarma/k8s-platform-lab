// Package repository provides persistence for admin-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/admin-service/internal/domain"
)

// Repository stores platform settings.
type Repository interface {
	Upsert(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (domain.Setting, error)
	List(ctx context.Context) ([]domain.Setting, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory setting store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Setting
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Setting{}} }

func (r *InMemory) Upsert(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key] = domain.Setting{Key: key, Value: value, UpdatedAt: time.Now().UTC()}
	return nil
}

func (r *InMemory) Get(_ context.Context, key string) (domain.Setting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.items[key]
	if !ok {
		return domain.Setting{}, apierror.ErrNotFound("setting not found")
	}
	return s, nil
}

func (r *InMemory) List(_ context.Context) ([]domain.Setting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Setting
	for _, s := range r.items {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// --- Postgres ---

// Postgres persists settings under the admin_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Upsert(ctx context.Context, key, value string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admin_settings (key, value, updated_at) VALUES ($1,$2,now())
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		key, value)
	return err
}

func (r *Postgres) Get(ctx context.Context, key string) (domain.Setting, error) {
	var s domain.Setting
	err := r.pool.QueryRow(ctx,
		`SELECT key, value, updated_at FROM admin_settings WHERE key=$1`, key).
		Scan(&s.Key, &s.Value, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Setting{}, apierror.ErrNotFound("setting not found")
	}
	if err != nil {
		return domain.Setting{}, err
	}
	return s, nil
}

func (r *Postgres) List(ctx context.Context) ([]domain.Setting, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value, updated_at FROM admin_settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Setting
	for rows.Next() {
		var s domain.Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
