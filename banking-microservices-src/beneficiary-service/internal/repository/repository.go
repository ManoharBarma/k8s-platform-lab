// Package repository provides persistence for beneficiary-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/beneficiary-service/internal/domain"
)

// Repository stores beneficiaries.
type Repository interface {
	Create(ctx context.Context, b domain.Beneficiary) error
	GetByID(ctx context.Context, id string) (domain.Beneficiary, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Beneficiary, error)
	Delete(ctx context.Context, id string) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory beneficiary store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Beneficiary
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Beneficiary{}} }

func (r *InMemory) Create(_ context.Context, b domain.Beneficiary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[b.ID] = b
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Beneficiary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.items[id]
	if !ok {
		return domain.Beneficiary{}, apierror.ErrNotFound("beneficiary not found")
	}
	return b, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.Beneficiary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Beneficiary
	for _, b := range r.items {
		if b.OwnerID == ownerID {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

// --- Postgres ---

// Postgres persists beneficiaries under the beneficiary_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, name, account_number, bank_name, routing_code, currency, created_at"

func (r *Postgres) Create(ctx context.Context, b domain.Beneficiary) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO beneficiary_beneficiaries (id, owner_id, name, account_number, bank_name, routing_code, currency, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.OwnerID, b.Name, b.AccountNumber, b.BankName, b.RoutingCode, b.Currency, b.CreatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Beneficiary, error) {
	var b domain.Beneficiary
	err := r.pool.QueryRow(ctx, `SELECT `+cols+` FROM beneficiary_beneficiaries WHERE id=$1`, id).
		Scan(&b.ID, &b.OwnerID, &b.Name, &b.AccountNumber, &b.BankName, &b.RoutingCode, &b.Currency, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Beneficiary{}, apierror.ErrNotFound("beneficiary not found")
	}
	return b, err
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.Beneficiary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM beneficiary_beneficiaries WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Beneficiary
	for rows.Next() {
		var b domain.Beneficiary
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Name, &b.AccountNumber, &b.BankName, &b.RoutingCode, &b.Currency, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Postgres) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM beneficiary_beneficiaries WHERE id=$1`, id)
	return err
}
