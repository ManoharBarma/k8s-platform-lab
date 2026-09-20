// Package repository provides persistence for currency-exchange-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/pkg/apierror"
	"banking-platform/services/currency-exchange-service/internal/domain"
)

// Repository stores exchange rates.
type Repository interface {
	Upsert(ctx context.Context, code string, perUSD decimal.Decimal) error
	Get(ctx context.Context, code string) (decimal.Decimal, error)
	List(ctx context.Context) ([]domain.Rate, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory rate store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]decimal.Decimal
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]decimal.Decimal{}} }

func (r *InMemory) Upsert(_ context.Context, code string, perUSD decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[code] = perUSD
	return nil
}

func (r *InMemory) Get(_ context.Context, code string) (decimal.Decimal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[code]
	if !ok {
		return decimal.Decimal{}, apierror.ErrNotFound("rate not found")
	}
	return v, nil
}

func (r *InMemory) List(_ context.Context) ([]domain.Rate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Rate
	for code, perUSD := range r.items {
		out = append(out, domain.Rate{Code: code, PerUSD: perUSD})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

// --- Postgres ---

// Postgres persists rates under the fx_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Upsert(ctx context.Context, code string, perUSD decimal.Decimal) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fx_rates (code, per_usd) VALUES ($1,$2)
		 ON CONFLICT (code) DO UPDATE SET per_usd = EXCLUDED.per_usd`,
		code, perUSD.String())
	return err
}

func (r *Postgres) Get(ctx context.Context, code string) (decimal.Decimal, error) {
	var perUSD string
	err := r.pool.QueryRow(ctx, `SELECT per_usd::text FROM fx_rates WHERE code=$1`, code).Scan(&perUSD)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Decimal{}, apierror.ErrNotFound("rate not found")
	}
	if err != nil {
		return decimal.Decimal{}, err
	}
	v, err := decimal.NewFromString(perUSD)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return v, nil
}

func (r *Postgres) List(ctx context.Context) ([]domain.Rate, error) {
	rows, err := r.pool.Query(ctx, `SELECT code, per_usd::text FROM fx_rates ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Rate
	for rows.Next() {
		var code, perUSD string
		if err := rows.Scan(&code, &perUSD); err != nil {
			return nil, err
		}
		v, err := decimal.NewFromString(perUSD)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.Rate{Code: code, PerUSD: v})
	}
	return out, rows.Err()
}
