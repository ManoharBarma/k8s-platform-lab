// Package repository provides persistence for loan-service.
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
	"banking-platform/services/loan-service/internal/domain"
)

// Repository stores loans.
type Repository interface {
	Create(ctx context.Context, l domain.Loan) error
	Update(ctx context.Context, l domain.Loan) error
	GetByID(ctx context.Context, id string) (domain.Loan, error)
	ListByBorrower(ctx context.Context, borrowerID string) ([]domain.Loan, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory loan store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Loan
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Loan{}} }

func (r *InMemory) Create(_ context.Context, l domain.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[l.ID] = l
	return nil
}

func (r *InMemory) Update(_ context.Context, l domain.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[l.ID]; !ok {
		return apierror.ErrNotFound("loan not found")
	}
	r.items[l.ID] = l
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.items[id]
	if !ok {
		return domain.Loan{}, apierror.ErrNotFound("loan not found")
	}
	return l, nil
}

func (r *InMemory) ListByBorrower(_ context.Context, borrowerID string) ([]domain.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Loan
	for _, l := range r.items {
		if l.BorrowerID == borrowerID {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists loans under the loan_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, borrower_id, account_id, principal::text, currency, annual_rate, tenure_months, emi_amount::text, outstanding::text, status, created_at, disbursed_at"

func (r *Postgres) Create(ctx context.Context, l domain.Loan) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO loan_loans
		 (id, borrower_id, account_id, principal, currency, annual_rate, tenure_months, emi_amount, outstanding, status, created_at)
		 VALUES ($1,$2,$3,$4::numeric,$5,$6,$7,$8::numeric,$9::numeric,$10,$11)`,
		l.ID, l.BorrowerID, l.AccountID, l.Principal.String(), l.Currency, l.AnnualRate,
		l.TenureMonths, l.EMIAmount.String(), l.Outstanding.String(), l.Status, l.CreatedAt)
	return err
}

func (r *Postgres) Update(ctx context.Context, l domain.Loan) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE loan_loans SET emi_amount=$2::numeric, outstanding=$3::numeric, status=$4, disbursed_at=$5 WHERE id=$1`,
		l.ID, l.EMIAmount.String(), l.Outstanding.String(), l.Status, l.DisbursedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("loan not found")
	}
	return nil
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Loan, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM loan_loans WHERE id=$1`, id))
}

func (r *Postgres) ListByBorrower(ctx context.Context, borrowerID string) ([]domain.Loan, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM loan_loans WHERE borrower_id=$1 ORDER BY created_at DESC`, borrowerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Loan
	for rows.Next() {
		l, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.Loan, error) {
	var l domain.Loan
	var principal, emi, outstanding string
	err := s.Scan(&l.ID, &l.BorrowerID, &l.AccountID, &principal, &l.Currency, &l.AnnualRate,
		&l.TenureMonths, &emi, &outstanding, &l.Status, &l.CreatedAt, &l.DisbursedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Loan{}, apierror.ErrNotFound("loan not found")
	}
	if err != nil {
		return domain.Loan{}, err
	}
	l.Principal, _ = decimal.NewFromString(principal)
	l.EMIAmount, _ = decimal.NewFromString(emi)
	l.Outstanding, _ = decimal.NewFromString(outstanding)
	return l, nil
}
