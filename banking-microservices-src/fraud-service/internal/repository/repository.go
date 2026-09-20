// Package repository provides persistence for fraud alerts.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/services/fraud-service/internal/domain"
)

// Repository is a store of fraud alerts.
type Repository interface {
	Create(ctx context.Context, a domain.Alert) error
	List(ctx context.Context, limit int) ([]domain.Alert, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory alert store.
type InMemory struct {
	mu     sync.RWMutex
	alerts []domain.Alert
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{} }

func (r *InMemory) Create(_ context.Context, a domain.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alerts = append(r.alerts, a)
	return nil
}

func (r *InMemory) List(_ context.Context, limit int) ([]domain.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Alert
	for i := len(r.alerts) - 1; i >= 0; i-- {
		out = append(out, r.alerts[i])
		if len(out) >= limit {
			break
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists fraud alerts under the fraud_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Create(ctx context.Context, a domain.Alert) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fraud_alerts (id, transaction_id, account_id, amount, currency, reason, score, created_at)
		 VALUES ($1,$2,$3,$4::numeric,$5,$6,$7,$8)`,
		a.ID, a.TransactionID, a.AccountID, a.Amount.String(), a.Currency, a.Reason, a.Score, a.CreatedAt)
	return err
}

func (r *Postgres) List(ctx context.Context, limit int) ([]domain.Alert, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, transaction_id, account_id, amount::text, currency, reason, score, created_at
		 FROM fraud_alerts ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Alert
	for rows.Next() {
		var a domain.Alert
		var amt string
		if err := rows.Scan(&a.ID, &a.TransactionID, &a.AccountID, &amt, &a.Currency, &a.Reason, &a.Score, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Amount, _ = decimal.NewFromString(amt)
		out = append(out, a)
	}
	return out, rows.Err()
}
