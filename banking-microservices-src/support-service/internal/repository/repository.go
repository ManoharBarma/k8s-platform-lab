// Package repository provides persistence for support-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/support-service/internal/domain"
)

// Repository stores support tickets.
type Repository interface {
	Create(ctx context.Context, t domain.Ticket) error
	GetByID(ctx context.Context, id string) (domain.Ticket, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Ticket, error)
	ListAll(ctx context.Context, limit int) ([]domain.Ticket, error)
	Update(ctx context.Context, t domain.Ticket) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory ticket store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Ticket
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Ticket{}} }

func (r *InMemory) Create(_ context.Context, t domain.Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return domain.Ticket{}, apierror.ErrNotFound("ticket not found")
	}
	return t, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Ticket
	for _, t := range r.items {
		if t.OwnerID == ownerID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) ListAll(_ context.Context, limit int) ([]domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Ticket
	for _, t := range r.items {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *InMemory) Update(_ context.Context, t domain.Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

// --- Postgres ---

// Postgres persists tickets under the support_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, subject, body, status, created_at, updated_at"

func (r *Postgres) Create(ctx context.Context, t domain.Ticket) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO support_tickets (id, owner_id, subject, body, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.OwnerID, t.Subject, t.Body, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Ticket, error) {
	var t domain.Ticket
	err := r.pool.QueryRow(ctx, `SELECT `+cols+` FROM support_tickets WHERE id=$1`, id).
		Scan(&t.ID, &t.OwnerID, &t.Subject, &t.Body, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ticket{}, apierror.ErrNotFound("ticket not found")
	}
	return t, err
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.Ticket, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM support_tickets WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.OwnerID, &t.Subject, &t.Body, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Postgres) ListAll(ctx context.Context, limit int) ([]domain.Ticket, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM support_tickets ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.OwnerID, &t.Subject, &t.Body, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Postgres) Update(ctx context.Context, t domain.Ticket) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE support_tickets SET owner_id=$2, subject=$3, body=$4, status=$5, created_at=$6, updated_at=$7 WHERE id=$1`,
		t.ID, t.OwnerID, t.Subject, t.Body, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}
