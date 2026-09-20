// Package repository provides persistence for investment-service.
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
	"banking-platform/services/investment-service/internal/domain"
)

// Repository stores holdings, unique per (owner, symbol).
type Repository interface {
	Get(ctx context.Context, ownerID, symbol string) (domain.Holding, error)
	Upsert(ctx context.Context, h domain.Holding) error
	Delete(ctx context.Context, ownerID, symbol string) error
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Holding, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory holding store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Holding // keyed by owner_id + "\x00" + symbol
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Holding{}} }

func key(ownerID, symbol string) string { return ownerID + "\x00" + symbol }

func (r *InMemory) Get(_ context.Context, ownerID, symbol string) (domain.Holding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.items[key(ownerID, symbol)]
	if !ok {
		return domain.Holding{}, apierror.ErrNotFound("holding not found")
	}
	return h, nil
}

func (r *InMemory) Upsert(_ context.Context, h domain.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key(h.OwnerID, h.Symbol)] = h
	return nil
}

func (r *InMemory) Delete(_ context.Context, ownerID, symbol string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, key(ownerID, symbol))
	return nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.Holding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Holding
	for _, h := range r.items {
		if h.OwnerID == ownerID {
			out = append(out, h)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists holdings under the inv_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, symbol, units::text, avg_price::text, updated_at"

func scanHolding(row pgx.Row) (domain.Holding, error) {
	var h domain.Holding
	var units, avgPrice string
	if err := row.Scan(&h.ID, &h.OwnerID, &h.Symbol, &units, &avgPrice, &h.UpdatedAt); err != nil {
		return domain.Holding{}, err
	}
	h.Units, _ = decimal.NewFromString(units)
	h.AvgPrice, _ = decimal.NewFromString(avgPrice)
	return h, nil
}

func (r *Postgres) Get(ctx context.Context, ownerID, symbol string) (domain.Holding, error) {
	h, err := scanHolding(r.pool.QueryRow(ctx,
		`SELECT `+cols+` FROM inv_holdings WHERE owner_id=$1 AND symbol=$2`, ownerID, symbol))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Holding{}, apierror.ErrNotFound("holding not found")
	}
	return h, err
}

func (r *Postgres) Upsert(ctx context.Context, h domain.Holding) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO inv_holdings (id, owner_id, symbol, units, avg_price, updated_at)
		 VALUES ($1,$2,$3,$4::numeric,$5::numeric,$6)
		 ON CONFLICT (owner_id, symbol) DO UPDATE
		 SET units = EXCLUDED.units, avg_price = EXCLUDED.avg_price, updated_at = EXCLUDED.updated_at`,
		h.ID, h.OwnerID, h.Symbol, h.Units.String(), h.AvgPrice.String(), h.UpdatedAt)
	return err
}

func (r *Postgres) Delete(ctx context.Context, ownerID, symbol string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM inv_holdings WHERE owner_id=$1 AND symbol=$2`, ownerID, symbol)
	return err
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.Holding, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM inv_holdings WHERE owner_id=$1 ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Holding
	for rows.Next() {
		h, err := scanHolding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
