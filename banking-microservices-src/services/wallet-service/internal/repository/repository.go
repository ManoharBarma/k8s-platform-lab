// Package repository persists wallets and their transactions, applying balance
// changes atomically (with an overdraft guard) and optionally an outbox event.
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
	"banking-platform/services/wallet-service/internal/domain"
)

// ErrInsufficientFunds is returned when a spend would overdraw the wallet.
var ErrInsufficientFunds = apierror.New(422, "insufficient_funds", "insufficient wallet balance")

// Repository stores wallets and transactions.
type Repository interface {
	Ensure(ctx context.Context, userID, currency string) (domain.Wallet, error)
	Get(ctx context.Context, userID string) (domain.Wallet, error)
	// Apply changes the balance by delta (signed) inside one transaction,
	// rejecting overdrafts, and records the transaction (+ optional event).
	Apply(ctx context.Context, t domain.Transaction, delta decimal.Decimal, event *outbox.Event) (domain.Wallet, error)
	TxnByKey(ctx context.Context, key string) (domain.Transaction, error)
	History(ctx context.Context, userID string) ([]domain.Transaction, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory wallet store.
type InMemory struct {
	mu      sync.Mutex
	wallets map[string]domain.Wallet
	txns    map[string]domain.Transaction // by id
	byKey   map[string]string             // idempotency key -> txn id
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{
		wallets: map[string]domain.Wallet{},
		txns:    map[string]domain.Transaction{},
		byKey:   map[string]string{},
	}
}

func (r *InMemory) Ensure(_ context.Context, userID, currency string) (domain.Wallet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.wallets[userID]
	if !ok {
		w = domain.Wallet{UserID: userID, Balance: decimal.Zero, Currency: currency}
		r.wallets[userID] = w
	}
	return w, nil
}

func (r *InMemory) Get(_ context.Context, userID string) (domain.Wallet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.wallets[userID]
	if !ok {
		return domain.Wallet{}, apierror.ErrNotFound("wallet not found")
	}
	return w, nil
}

func (r *InMemory) Apply(_ context.Context, t domain.Transaction, delta decimal.Decimal, _ *outbox.Event) (domain.Wallet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.wallets[t.UserID]
	if !ok {
		return domain.Wallet{}, apierror.ErrNotFound("wallet not found")
	}
	newBal := w.Balance.Add(delta)
	if newBal.IsNegative() {
		return domain.Wallet{}, ErrInsufficientFunds
	}
	w.Balance = newBal
	w.UpdatedAt = t.CreatedAt
	r.wallets[t.UserID] = w
	r.txns[t.ID] = t
	r.byKey[t.IdempotencyKey] = t.ID
	return w, nil
}

func (r *InMemory) TxnByKey(_ context.Context, key string) (domain.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byKey[key]
	if !ok {
		return domain.Transaction{}, apierror.ErrNotFound("transaction not found")
	}
	return r.txns[id], nil
}

func (r *InMemory) History(_ context.Context, userID string) ([]domain.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Transaction
	for _, t := range r.txns {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists wallets under the wallet_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Ensure(ctx context.Context, userID, currency string) (domain.Wallet, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO wallet_wallets (user_id, currency, balance) VALUES ($1,$2,0)
		 ON CONFLICT (user_id) DO NOTHING`, userID, currency)
	if err != nil {
		return domain.Wallet{}, err
	}
	return r.Get(ctx, userID)
}

func (r *Postgres) Get(ctx context.Context, userID string) (domain.Wallet, error) {
	var w domain.Wallet
	var bal string
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, balance::text, currency, updated_at FROM wallet_wallets WHERE user_id=$1`, userID).
		Scan(&w.UserID, &bal, &w.Currency, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, apierror.ErrNotFound("wallet not found")
	}
	if err != nil {
		return domain.Wallet{}, err
	}
	w.Balance, _ = decimal.NewFromString(bal)
	return w, nil
}

func (r *Postgres) Apply(ctx context.Context, t domain.Transaction, delta decimal.Decimal, event *outbox.Event) (domain.Wallet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Wallet{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Overdraft-safe balance update: only applies when the result stays >= 0.
	var newBal string
	err = tx.QueryRow(ctx,
		`UPDATE wallet_wallets SET balance = balance + $2::numeric, updated_at = now()
		 WHERE user_id = $1 AND balance + $2::numeric >= 0
		 RETURNING balance::text`, t.UserID, delta.String()).Scan(&newBal)
	if errors.Is(err, pgx.ErrNoRows) {
		// Either the wallet doesn't exist or the guard failed (insufficient).
		var exists bool
		if e := tx.QueryRow(ctx, `SELECT true FROM wallet_wallets WHERE user_id=$1`, t.UserID).Scan(&exists); e != nil {
			return domain.Wallet{}, apierror.ErrNotFound("wallet not found")
		}
		return domain.Wallet{}, ErrInsufficientFunds
	}
	if err != nil {
		return domain.Wallet{}, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO wallet_transactions (id, user_id, type, amount, idempotency_key, created_at)
		 VALUES ($1,$2,$3,$4::numeric,$5,$6)`,
		t.ID, t.UserID, t.Type, t.Amount.String(), t.IdempotencyKey, t.CreatedAt); err != nil {
		return domain.Wallet{}, err
	}
	if event != nil {
		if err := outbox.Insert(ctx, tx, *event); err != nil {
			return domain.Wallet{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Wallet{}, err
	}

	bal, _ := decimal.NewFromString(newBal)
	return domain.Wallet{UserID: t.UserID, Balance: bal, Currency: "", UpdatedAt: t.CreatedAt}, nil
}

func (r *Postgres) TxnByKey(ctx context.Context, key string) (domain.Transaction, error) {
	var t domain.Transaction
	var amt string
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, type, amount::text, idempotency_key, created_at
		 FROM wallet_transactions WHERE idempotency_key=$1`, key).
		Scan(&t.ID, &t.UserID, &t.Type, &amt, &t.IdempotencyKey, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, apierror.ErrNotFound("transaction not found")
	}
	if err != nil {
		return domain.Transaction{}, err
	}
	t.Amount, _ = decimal.NewFromString(amt)
	return t, nil
}

func (r *Postgres) History(ctx context.Context, userID string) ([]domain.Transaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, amount::text, idempotency_key, created_at
		 FROM wallet_transactions WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var amt string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &amt, &t.IdempotencyKey, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Amount, _ = decimal.NewFromString(amt)
		out = append(out, t)
	}
	return out, rows.Err()
}
