// Package repository provides persistence for analytics metrics.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/services/analytics-service/internal/domain"
)

// Repository is a store of per-transaction-type aggregates.
type Repository interface {
	// Add increments the count by one and adds amount to the running sum for key.
	Add(ctx context.Context, key string, amount decimal.Decimal) error
	List(ctx context.Context) ([]domain.Metric, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory metric store.
type InMemory struct {
	mu      sync.RWMutex
	metrics map[string]domain.Metric
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{metrics: map[string]domain.Metric{}}
}

func (r *InMemory) Add(_ context.Context, key string, amount decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.metrics[key]
	if !ok {
		m = domain.Metric{Key: key, Count: 0, Sum: decimal.Zero}
	}
	m.Count++
	m.Sum = m.Sum.Add(amount)
	r.metrics[key] = m
	return nil
}

func (r *InMemory) List(_ context.Context) ([]domain.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// --- Postgres ---

// Postgres persists metrics under the analytics_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Add(ctx context.Context, key string, amount decimal.Decimal) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO analytics_metrics (key, count, sum) VALUES ($1, 1, $2::numeric)
		 ON CONFLICT (key) DO UPDATE
		 SET count = analytics_metrics.count + 1,
		     sum = analytics_metrics.sum + EXCLUDED.sum`,
		key, amount.String())
	return err
}

func (r *Postgres) List(ctx context.Context) ([]domain.Metric, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT key, count, sum::text FROM analytics_metrics ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Metric
	for rows.Next() {
		var m domain.Metric
		var sum string
		if err := rows.Scan(&m.Key, &m.Count, &sum); err != nil {
			return nil, err
		}
		m.Sum, _ = decimal.NewFromString(sum)
		out = append(out, m)
	}
	return out, rows.Err()
}
