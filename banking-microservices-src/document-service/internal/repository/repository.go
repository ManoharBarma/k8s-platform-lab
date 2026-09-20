// Package repository provides persistence for document metadata.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/document-service/internal/domain"
)

// Repository stores document metadata.
type Repository interface {
	Create(ctx context.Context, d domain.Document) error
	GetByID(ctx context.Context, id string) (domain.Document, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Document, error)
	Delete(ctx context.Context, id string) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory document store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Document
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Document{}} }

func (r *InMemory) Create(_ context.Context, d domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[d.ID] = d
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.items[id]
	if !ok {
		return domain.Document{}, apierror.ErrNotFound("document not found")
	}
	return d, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Document
	for _, d := range r.items {
		if d.OwnerID == ownerID {
			out = append(out, d)
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

// Postgres persists document metadata under the document_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const docCols = "id, owner_id, type, key, content_type, created_at"

func (r *Postgres) Create(ctx context.Context, d domain.Document) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO document_documents (id, owner_id, type, key, content_type, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		d.ID, d.OwnerID, d.Type, d.Key, d.ContentType, d.CreatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Document, error) {
	var d domain.Document
	err := r.pool.QueryRow(ctx, `SELECT `+docCols+` FROM document_documents WHERE id=$1`, id).
		Scan(&d.ID, &d.OwnerID, &d.Type, &d.Key, &d.ContentType, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Document{}, apierror.ErrNotFound("document not found")
	}
	return d, err
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.Document, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+docCols+` FROM document_documents WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Document
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.OwnerID, &d.Type, &d.Key, &d.ContentType, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Postgres) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM document_documents WHERE id=$1`, id)
	return err
}
