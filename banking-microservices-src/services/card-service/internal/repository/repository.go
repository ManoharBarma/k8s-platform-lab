// Package repository provides persistence for card-service.
package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/card-service/internal/domain"
)

// Repository stores cards.
type Repository interface {
	Create(ctx context.Context, c domain.Card) error
	GetByID(ctx context.Context, id string) (domain.Card, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Card, error)
	Update(ctx context.Context, c domain.Card) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory card store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Card
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Card{}} }

func (r *InMemory) Create(_ context.Context, c domain.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[c.ID] = c
	return nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (domain.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.items[id]
	if !ok {
		return domain.Card{}, apierror.ErrNotFound("card not found")
	}
	return c, nil
}

func (r *InMemory) ListByOwner(_ context.Context, ownerID string) ([]domain.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Card
	for _, c := range r.items {
		if c.OwnerID == ownerID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) Update(_ context.Context, c domain.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[c.ID] = c
	return nil
}

// --- Postgres ---

// Postgres persists cards under the card_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = "id, owner_id, account_id, type, network, masked_number, status, expiry_month, expiry_year, created_at"

func (r *Postgres) Create(ctx context.Context, c domain.Card) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO card_cards (id, owner_id, account_id, type, network, masked_number, status, expiry_month, expiry_year, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.OwnerID, c.AccountID, c.Type, c.Network, c.MaskedNumber, c.Status, c.ExpiryMonth, c.ExpiryYear, c.CreatedAt)
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Card, error) {
	var c domain.Card
	err := r.pool.QueryRow(ctx, `SELECT `+cols+` FROM card_cards WHERE id=$1`, id).
		Scan(&c.ID, &c.OwnerID, &c.AccountID, &c.Type, &c.Network, &c.MaskedNumber, &c.Status, &c.ExpiryMonth, &c.ExpiryYear, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Card{}, apierror.ErrNotFound("card not found")
	}
	return c, err
}

func (r *Postgres) ListByOwner(ctx context.Context, ownerID string) ([]domain.Card, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM card_cards WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Card
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.AccountID, &c.Type, &c.Network, &c.MaskedNumber, &c.Status, &c.ExpiryMonth, &c.ExpiryYear, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Postgres) Update(ctx context.Context, c domain.Card) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE card_cards SET owner_id=$2, account_id=$3, type=$4, network=$5, masked_number=$6, status=$7, expiry_month=$8, expiry_year=$9, created_at=$10 WHERE id=$1`,
		c.ID, c.OwnerID, c.AccountID, c.Type, c.Network, c.MaskedNumber, c.Status, c.ExpiryMonth, c.ExpiryYear, c.CreatedAt)
	return err
}
