// Package repository provides persistence for account-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/account-service/internal/domain"
)

// Repository stores bank accounts.
type Repository interface {
	Create(ctx context.Context, a domain.Account) error
	GetByID(ctx context.Context, id string) (domain.Account, error)
	ListByCustomer(ctx context.Context, customerID string) ([]domain.Account, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory account store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Account
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Account{}} }

func (r *InMemory) Create(_ context.Context, a domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[a.ID] = a
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.items[id]
	if !ok {
		return domain.Account{}, apierror.ErrNotFound("account not found")
	}
	return a, nil
}

func (r *InMemory) ListByCustomer(_ context.Context, customerID string) ([]domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Account
	for _, a := range r.items {
		if a.CustomerID == customerID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists accounts under the account_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const acctCols = "id, account_number, customer_id, type, currency, status, created_at"

func (r *Postgres) Create(ctx context.Context, a domain.Account) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO account_accounts (id, account_number, customer_id, type, currency, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.AccountNumber, a.CustomerID, a.Type, a.Currency, a.Status, a.CreatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Account, error) {
	var a domain.Account
	err := r.pool.QueryRow(ctx, `SELECT `+acctCols+` FROM account_accounts WHERE id=$1`, id).
		Scan(&a.ID, &a.AccountNumber, &a.CustomerID, &a.Type, &a.Currency, &a.Status, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, apierror.ErrNotFound("account not found")
	}
	return a, err
}

func (r *Postgres) ListByCustomer(ctx context.Context, customerID string) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+acctCols+` FROM account_accounts WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.AccountNumber, &a.CustomerID, &a.Type, &a.Currency, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
