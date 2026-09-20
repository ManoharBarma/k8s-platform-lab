// Package repository persists payments and (transactionally) outbox events.
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
	"banking-platform/pkg/outbox"
	"banking-platform/services/payment-service/internal/domain"
)

// Repository stores payments.
type Repository interface {
	Create(ctx context.Context, p domain.Payment, event *outbox.Event) error
	GetByID(ctx context.Context, id string) (domain.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Payment, error)
	ListByPayer(ctx context.Context, payerAccountID string) ([]domain.Payment, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory payment store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Payment
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Payment{}} }

func (r *InMemory) Create(_ context.Context, p domain.Payment, _ *outbox.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.ID] = p
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return domain.Payment{}, apierror.ErrNotFound("payment not found")
	}
	return p, nil
}

func (r *InMemory) GetByIdempotencyKey(_ context.Context, key string) (domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.IdempotencyKey == key {
			return p, nil
		}
	}
	return domain.Payment{}, apierror.ErrNotFound("payment not found")
}

func (r *InMemory) ListByPayer(_ context.Context, payerAccountID string) ([]domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Payment
	for _, p := range r.items {
		if p.PayerAccountID == payerAccountID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists payments under the payment_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, payer_account_id, beneficiary_id, amount::text, currency, status, reference, idempotency_key, created_at"

func (r *Postgres) Create(ctx context.Context, p domain.Payment, event *outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`INSERT INTO payment_payments
		 (id, payer_account_id, beneficiary_id, amount, currency, status, reference, idempotency_key, created_at)
		 VALUES ($1,$2,$3,$4::numeric,$5,$6,$7,$8,$9)`,
		p.ID, p.PayerAccountID, p.BeneficiaryID, p.Amount.String(), p.Currency, p.Status,
		p.Reference, p.IdempotencyKey, p.CreatedAt); err != nil {
		return err
	}
	if event != nil {
		if err := outbox.Insert(ctx, tx, *event); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Payment, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM payment_payments WHERE id=$1`, id))
}

func (r *Postgres) GetByIdempotencyKey(ctx context.Context, key string) (domain.Payment, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM payment_payments WHERE idempotency_key=$1`, key))
}

func (r *Postgres) ListByPayer(ctx context.Context, payerAccountID string) ([]domain.Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM payment_payments WHERE payer_account_id=$1 ORDER BY created_at DESC`, payerAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Payment
	for rows.Next() {
		p, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.Payment, error) {
	var p domain.Payment
	var amt string
	err := s.Scan(&p.ID, &p.PayerAccountID, &p.BeneficiaryID, &amt, &p.Currency, &p.Status,
		&p.Reference, &p.IdempotencyKey, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, apierror.ErrNotFound("payment not found")
	}
	if err != nil {
		return domain.Payment{}, err
	}
	p.Amount, _ = decimal.NewFromString(amt)
	return p, nil
}
