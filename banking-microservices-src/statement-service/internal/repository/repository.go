// Package repository provides persistence for statement entries.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/services/statement-service/internal/domain"
)

// Repository is a store of statement entries.
type Repository interface {
	CreateMany(ctx context.Context, entries []domain.Entry) error
	ListByAccount(ctx context.Context, accountID string, limit int) ([]domain.Entry, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory statement store.
type InMemory struct {
	mu      sync.RWMutex
	entries []domain.Entry
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{} }

func (r *InMemory) CreateMany(_ context.Context, entries []domain.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entries...)
	return nil
}

func (r *InMemory) ListByAccount(_ context.Context, accountID string, limit int) ([]domain.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Entry
	for i := len(r.entries) - 1; i >= 0; i-- {
		if r.entries[i].AccountID == accountID {
			out = append(out, r.entries[i])
			if len(out) >= limit {
				break
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists statement entries under the statement_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) CreateMany(ctx context.Context, entries []domain.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, e := range entries {
		if _, err := tx.Exec(ctx,
			`INSERT INTO statement_entries (id, account_id, transaction_id, type, direction, amount, currency, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8)`,
			e.ID, e.AccountID, e.TransactionID, e.Type, e.Direction, e.Amount.String(), e.Currency, e.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Postgres) ListByAccount(ctx context.Context, accountID string, limit int) ([]domain.Entry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, account_id, transaction_id, type, direction, amount::text, currency, created_at
		 FROM statement_entries WHERE account_id=$1 ORDER BY created_at DESC LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Entry
	for rows.Next() {
		var e domain.Entry
		var amt string
		if err := rows.Scan(&e.ID, &e.AccountID, &e.TransactionID, &e.Type, &e.Direction, &amt, &e.Currency, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Amount, _ = decimal.NewFromString(amt)
		out = append(out, e)
	}
	return out, rows.Err()
}
