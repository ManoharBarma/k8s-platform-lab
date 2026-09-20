// Package repository provides persistence for fixed-deposit-service.
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
	"banking-platform/services/fixed-deposit-service/internal/domain"
)

// Repository stores fixed deposits.
type Repository interface {
	Create(ctx context.Context, fd domain.FixedDeposit) error
	GetByID(ctx context.Context, id string) (domain.FixedDeposit, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.FixedDeposit, error)
	Update(ctx context.Context, fd domain.FixedDeposit) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory fixed deposit store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.FixedDeposit
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.FixedDeposit{}} }

func (r *InMemory) Create(_ context.Context, fd domain.FixedDeposit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[fd.ID] = fd
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.FixedDeposit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fd, ok := r.items[id]
	if !ok {
		return domain.FixedDeposit{}, apierror.ErrNotFound("fixed deposit not found")
	}
	return fd, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.FixedDeposit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.FixedDeposit
	for _, fd := range r.items {
		if fd.OwnerID == ownerID {
			out = append(out, fd)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) Update(_ context.Context, fd domain.FixedDeposit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[fd.ID]; !ok {
		return apierror.ErrNotFound("fixed deposit not found")
	}
	r.items[fd.ID] = fd
	return nil
}

// --- Postgres ---

// Postgres persists fixed deposits under the fd_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, principal::text, currency, annual_rate_pct, tenure_months, maturity_amount::text, status, created_at, matures_at"

func (r *Postgres) Create(ctx context.Context, fd domain.FixedDeposit) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fd_deposits
		 (id, owner_id, principal, currency, annual_rate_pct, tenure_months, maturity_amount, status, created_at, matures_at)
		 VALUES ($1,$2,$3::numeric,$4,$5,$6,$7::numeric,$8,$9,$10)`,
		fd.ID, fd.OwnerID, fd.Principal.String(), fd.Currency, fd.AnnualRatePct,
		fd.TenureMonths, fd.MaturityAmount.String(), fd.Status, fd.CreatedAt, fd.MaturesAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.FixedDeposit, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM fd_deposits WHERE id=$1`, id))
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.FixedDeposit, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM fd_deposits WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, fd)
	}
	return out, rows.Err()
}

func (r *Postgres) Update(ctx context.Context, fd domain.FixedDeposit) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE fd_deposits SET maturity_amount=$2::numeric, status=$3, matures_at=$4 WHERE id=$1`,
		fd.ID, fd.MaturityAmount.String(), fd.Status, fd.MaturesAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("fixed deposit not found")
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.FixedDeposit, error) {
	var fd domain.FixedDeposit
	var principal, maturity string
	err := s.Scan(&fd.ID, &fd.OwnerID, &principal, &fd.Currency, &fd.AnnualRatePct,
		&fd.TenureMonths, &maturity, &fd.Status, &fd.CreatedAt, &fd.MaturesAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FixedDeposit{}, apierror.ErrNotFound("fixed deposit not found")
	}
	if err != nil {
		return domain.FixedDeposit{}, err
	}
	fd.Principal, _ = decimal.NewFromString(principal)
	fd.MaturityAmount, _ = decimal.NewFromString(maturity)
	return fd, nil
}
