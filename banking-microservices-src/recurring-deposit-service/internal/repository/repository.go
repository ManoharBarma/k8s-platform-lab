// Package repository provides persistence for recurring-deposit-service.
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
	"banking-platform/services/recurring-deposit-service/internal/domain"
)

// Repository stores recurring deposits.
type Repository interface {
	Create(ctx context.Context, d domain.RecurringDeposit) error
	GetByID(ctx context.Context, id string) (domain.RecurringDeposit, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.RecurringDeposit, error)
	Update(ctx context.Context, d domain.RecurringDeposit) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory recurring deposit store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.RecurringDeposit
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.RecurringDeposit{}} }

func (r *InMemory) Create(_ context.Context, d domain.RecurringDeposit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[d.ID] = d
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.RecurringDeposit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.items[id]
	if !ok {
		return domain.RecurringDeposit{}, apierror.ErrNotFound("recurring deposit not found")
	}
	return d, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.RecurringDeposit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.RecurringDeposit
	for _, d := range r.items {
		if d.OwnerID == ownerID {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) Update(_ context.Context, d domain.RecurringDeposit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[d.ID]; !ok {
		return apierror.ErrNotFound("recurring deposit not found")
	}
	r.items[d.ID] = d
	return nil
}

// --- Postgres ---

// Postgres persists recurring deposits under the rd_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, monthly_amount::text, currency, annual_rate_pct, tenure_months, installments_paid, maturity_amount::text, status, created_at"

func (r *Postgres) Create(ctx context.Context, d domain.RecurringDeposit) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rd_deposits
		 (id, owner_id, monthly_amount, currency, annual_rate_pct, tenure_months, installments_paid, maturity_amount, status, created_at)
		 VALUES ($1,$2,$3::numeric,$4,$5,$6,$7,$8::numeric,$9,$10)`,
		d.ID, d.OwnerID, d.MonthlyAmount.String(), d.Currency, d.AnnualRatePct,
		d.TenureMonths, d.InstallmentsPaid, d.MaturityAmount.String(), d.Status, d.CreatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.RecurringDeposit, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM rd_deposits WHERE id=$1`, id))
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.RecurringDeposit, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM rd_deposits WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RecurringDeposit
	for rows.Next() {
		d, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Postgres) Update(ctx context.Context, d domain.RecurringDeposit) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE rd_deposits SET installments_paid=$2, maturity_amount=$3::numeric, status=$4 WHERE id=$1`,
		d.ID, d.InstallmentsPaid, d.MaturityAmount.String(), d.Status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("recurring deposit not found")
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.RecurringDeposit, error) {
	var d domain.RecurringDeposit
	var monthly, maturity string
	err := s.Scan(&d.ID, &d.OwnerID, &monthly, &d.Currency, &d.AnnualRatePct,
		&d.TenureMonths, &d.InstallmentsPaid, &maturity, &d.Status, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RecurringDeposit{}, apierror.ErrNotFound("recurring deposit not found")
	}
	if err != nil {
		return domain.RecurringDeposit{}, err
	}
	d.MonthlyAmount, _ = decimal.NewFromString(monthly)
	d.MaturityAmount, _ = decimal.NewFromString(maturity)
	return d, nil
}
