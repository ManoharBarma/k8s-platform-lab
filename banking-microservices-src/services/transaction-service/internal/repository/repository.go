// Package repository persists transaction records and (atomically) outbox
// events.
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
	"banking-platform/services/transaction-service/internal/domain"
)

// Repository stores transaction records. When an outbox event is supplied it is
// written in the SAME transaction as the record (transactional outbox).
type Repository interface {
	Create(ctx context.Context, t domain.Transaction, event *outbox.Event) error
	GetByID(ctx context.Context, id string) (domain.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, error)
	ListByAccount(ctx context.Context, accountID string) ([]domain.Transaction, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory store (outbox events are dropped).
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Transaction
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Transaction{}} }

func (r *InMemory) Create(_ context.Context, t domain.Transaction, _ *outbox.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return domain.Transaction{}, apierror.ErrNotFound("transaction not found")
	}
	return t, nil
}

func (r *InMemory) GetByIdempotencyKey(_ context.Context, key string) (domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.items {
		if t.IdempotencyKey == key {
			return t, nil
		}
	}
	return domain.Transaction{}, apierror.ErrNotFound("transaction not found")
}

func (r *InMemory) ListByAccount(_ context.Context, accountID string) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Transaction
	for _, t := range r.items {
		if t.FromAccountID == accountID || t.ToAccountID == accountID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists transactions under the transaction_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

// amount is cast to text so it scans cleanly into a string for decimal parsing.
const txnCols = "id, idempotency_key, type, from_account_id, to_account_id, amount::text, currency, status, reference, created_at"

func (r *Postgres) Create(ctx context.Context, t domain.Transaction, event *outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO transaction_transactions
		 (id, idempotency_key, type, from_account_id, to_account_id, amount, currency, status, reference, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8,$9,$10)`,
		t.ID, t.IdempotencyKey, t.Type, nullify(t.FromAccountID), nullify(t.ToAccountID),
		t.Amount.String(), t.Currency, t.Status, t.Reference, t.CreatedAt)
	if err != nil {
		return err
	}
	if event != nil {
		if err := outbox.Insert(ctx, tx, *event); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Transaction, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+txnCols+` FROM transaction_transactions WHERE id=$1`, id))
}

func (r *Postgres) GetByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+txnCols+` FROM transaction_transactions WHERE idempotency_key=$1`, key))
}

func (r *Postgres) ListByAccount(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+txnCols+` FROM transaction_transactions
		 WHERE from_account_id=$1 OR to_account_id=$1 ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Transaction
	for rows.Next() {
		t, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.Transaction, error) {
	var t domain.Transaction
	var from, to *string
	var amount string
	err := s.Scan(&t.ID, &t.IdempotencyKey, &t.Type, &from, &to, &amount,
		&t.Currency, &t.Status, &t.Reference, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, apierror.ErrNotFound("transaction not found")
	}
	if err != nil {
		return domain.Transaction{}, err
	}
	if from != nil {
		t.FromAccountID = *from
	}
	if to != nil {
		t.ToAccountID = *to
	}
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return domain.Transaction{}, err
	}
	t.Amount = amt
	return t, nil
}

func nullify(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
