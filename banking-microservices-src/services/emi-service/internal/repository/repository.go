// Package repository provides persistence for emi-service.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/pkg/apierror"
	"banking-platform/services/emi-service/internal/domain"
)

// Repository stores amortization schedules.
type Repository interface {
	SaveSchedule(ctx context.Context, loanID string, installments []domain.Installment) error
	ListByLoan(ctx context.Context, loanID string) ([]domain.Installment, error)
	MarkPaid(ctx context.Context, installmentID string) error
	HasSchedule(ctx context.Context, loanID string) (bool, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory schedule store.
type InMemory struct {
	mu    sync.RWMutex
	byID  map[string]domain.Installment
	byLon map[string][]string
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{byID: map[string]domain.Installment{}, byLon: map[string][]string{}}
}

func (r *InMemory) SaveSchedule(_ context.Context, loanID string, installments []domain.Installment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ins := range installments {
		r.byID[ins.ID] = ins
		r.byLon[loanID] = append(r.byLon[loanID], ins.ID)
	}
	return nil
}

func (r *InMemory) ListByLoan(_ context.Context, loanID string) ([]domain.Installment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Installment
	for _, id := range r.byLon[loanID] {
		out = append(out, r.byID[id])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (r *InMemory) MarkPaid(_ context.Context, installmentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ins, ok := r.byID[installmentID]
	if !ok {
		return apierror.ErrNotFound("installment not found")
	}
	ins.Paid = true
	r.byID[installmentID] = ins
	return nil
}

func (r *InMemory) HasSchedule(_ context.Context, loanID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byLon[loanID]) > 0, nil
}

// --- Postgres ---

// Postgres persists installments under the emi_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) SaveSchedule(ctx context.Context, loanID string, installments []domain.Installment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, ins := range installments {
		if _, err := tx.Exec(ctx,
			`INSERT INTO emi_installments (id, loan_id, number, due_date, principal, interest, total, balance, paid)
			 VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8::numeric,false)`,
			ins.ID, ins.LoanID, ins.Number, ins.DueDate, ins.Principal.String(),
			ins.Interest.String(), ins.Total.String(), ins.Balance.String()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Postgres) ListByLoan(ctx context.Context, loanID string) ([]domain.Installment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, loan_id, number, due_date, principal::text, interest::text, total::text, balance::text, paid, paid_at
		 FROM emi_installments WHERE loan_id=$1 ORDER BY number`, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Installment
	for rows.Next() {
		var ins domain.Installment
		var principal, interest, total, balance string
		if err := rows.Scan(&ins.ID, &ins.LoanID, &ins.Number, &ins.DueDate, &principal, &interest,
			&total, &balance, &ins.Paid, &ins.PaidAt); err != nil {
			return nil, err
		}
		ins.Principal, _ = decimal.NewFromString(principal)
		ins.Interest, _ = decimal.NewFromString(interest)
		ins.Total, _ = decimal.NewFromString(total)
		ins.Balance, _ = decimal.NewFromString(balance)
		out = append(out, ins)
	}
	return out, rows.Err()
}

func (r *Postgres) MarkPaid(ctx context.Context, installmentID string) error {
	ct, err := r.pool.Exec(ctx, `UPDATE emi_installments SET paid=true, paid_at=now() WHERE id=$1`, installmentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("installment not found")
	}
	return nil
}

func (r *Postgres) HasSchedule(ctx context.Context, loanID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM emi_installments WHERE loan_id=$1)`, loanID).Scan(&exists)
	return exists, err
}
