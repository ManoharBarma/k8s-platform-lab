// Package repository provides persistence for the search index.
package repository

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/services/search-service/internal/domain"
)

// Repository is a store of searchable index documents.
type Repository interface {
	Index(ctx context.Context, d domain.Document) error
	Search(ctx context.Context, query string, limit int) ([]domain.Document, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory search index.
type InMemory struct {
	mu   sync.RWMutex
	docs []domain.Document
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{} }

func (r *InMemory) Index(_ context.Context, d domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs = append(r.docs, d)
	return nil
}

func (r *InMemory) Search(_ context.Context, query string, limit int) ([]domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := strings.ToLower(query)
	var out []domain.Document
	for i := len(r.docs) - 1; i >= 0; i-- {
		if strings.Contains(r.docs[i].Content, q) {
			out = append(out, r.docs[i])
			if len(out) >= limit {
				break
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// --- Postgres ---

// Postgres persists index documents under the search_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Index(ctx context.Context, d domain.Document) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO search_index (id, ref_id, kind, content, created_at) VALUES ($1,$2,$3,$4,$5)`,
		d.ID, d.RefID, d.Kind, d.Content, d.CreatedAt)
	return err
}

func (r *Postgres) Search(ctx context.Context, query string, limit int) ([]domain.Document, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, ref_id, kind, content, created_at FROM search_index WHERE content ILIKE '%'||$1||'%' ORDER BY created_at DESC LIMIT $2`,
		query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Document
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.RefID, &d.Kind, &d.Content, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
