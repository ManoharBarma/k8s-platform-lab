// Package repository provides persistence for kyc-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/kyc-service/internal/domain"
)

// Repository stores KYC verifications.
type Repository interface {
	Create(ctx context.Context, v domain.Verification) error
	GetByID(ctx context.Context, id string) (domain.Verification, error)
	LatestByUser(ctx context.Context, userID string) (domain.Verification, error)
	UpdateReview(ctx context.Context, v domain.Verification) error
	ListByStatus(ctx context.Context, status domain.Status) ([]domain.Verification, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory KYC store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Verification
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Verification{}} }

func (r *InMemory) Create(_ context.Context, v domain.Verification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[v.ID] = v
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Verification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[id]
	if !ok {
		return domain.Verification{}, apierror.ErrNotFound("verification not found")
	}
	return v, nil
}

func (r *InMemory) LatestByUser(_ context.Context, userID string) (domain.Verification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest domain.Verification
	found := false
	for _, v := range r.items {
		if v.UserID == userID && (!found || v.SubmittedAt.After(latest.SubmittedAt)) {
			latest, found = v, true
		}
	}
	if !found {
		return domain.Verification{}, apierror.ErrNotFound("verification not found")
	}
	return latest, nil
}

func (r *InMemory) UpdateReview(_ context.Context, v domain.Verification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[v.ID]; !ok {
		return apierror.ErrNotFound("verification not found")
	}
	r.items[v.ID] = v
	return nil
}

func (r *InMemory) ListByStatus(_ context.Context, status domain.Status) ([]domain.Verification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Verification
	for _, v := range r.items {
		if v.Status == status {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SubmittedAt.Before(out[j].SubmittedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists KYC verifications under the kyc_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const kycCols = "id, user_id, document_type, document_id, status, reason, submitted_at, reviewed_at"

func (r *Postgres) Create(ctx context.Context, v domain.Verification) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO kyc_verifications (id, user_id, document_type, document_id, status, reason, submitted_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		v.ID, v.UserID, v.DocumentType, v.DocumentID, v.Status, v.Reason, v.SubmittedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Verification, error) {
	return scanOne(r.pool.QueryRow(ctx, `SELECT `+kycCols+` FROM kyc_verifications WHERE id = $1`, id))
}

func (r *Postgres) LatestByUser(ctx context.Context, userID string) (domain.Verification, error) {
	return scanOne(r.pool.QueryRow(ctx,
		`SELECT `+kycCols+` FROM kyc_verifications WHERE user_id = $1 ORDER BY submitted_at DESC LIMIT 1`, userID))
}

func (r *Postgres) UpdateReview(ctx context.Context, v domain.Verification) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE kyc_verifications SET status=$2, reason=$3, reviewed_at=$4 WHERE id=$1`,
		v.ID, v.Status, v.Reason, v.ReviewedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("verification not found")
	}
	return nil
}

func (r *Postgres) ListByStatus(ctx context.Context, status domain.Status) ([]domain.Verification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+kycCols+` FROM kyc_verifications WHERE status=$1 ORDER BY submitted_at ASC`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Verification
	for rows.Next() {
		v, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanOne(s scanner) (domain.Verification, error) {
	var v domain.Verification
	err := s.Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentID, &v.Status, &v.Reason, &v.SubmittedAt, &v.ReviewedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Verification{}, apierror.ErrNotFound("verification not found")
	}
	return v, err
}
