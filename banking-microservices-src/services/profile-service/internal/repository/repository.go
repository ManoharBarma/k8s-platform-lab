// Package repository provides persistence for profile-service (in-memory +
// Postgres implementations of a simple upsert-by-user store).
package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/profile-service/internal/domain"
)

// Repository stores profiles keyed by user ID.
type Repository interface {
	Upsert(ctx context.Context, p domain.Profile) error
	Get(ctx context.Context, userID string) (domain.Profile, error)
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory profile store.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Profile
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{items: map[string]domain.Profile{}} }

func (r *InMemory) Upsert(_ context.Context, p domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.UserID] = p
	return nil
}

func (r *InMemory) Get(_ context.Context, userID string) (domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[userID]
	if !ok {
		return domain.Profile{}, apierror.ErrNotFound("profile not found")
	}
	return p, nil
}

// --- Postgres ---

// Postgres persists profiles under the profile_ prefix in the shared database.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Upsert(ctx context.Context, p domain.Profile) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO profile_profiles
		 (user_id, date_of_birth, gender, address_line1, address_line2, city, state, country, postal_code, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (user_id) DO UPDATE SET
		   date_of_birth=EXCLUDED.date_of_birth, gender=EXCLUDED.gender,
		   address_line1=EXCLUDED.address_line1, address_line2=EXCLUDED.address_line2,
		   city=EXCLUDED.city, state=EXCLUDED.state, country=EXCLUDED.country,
		   postal_code=EXCLUDED.postal_code, updated_at=EXCLUDED.updated_at`,
		p.UserID, p.DateOfBirth, p.Gender, p.AddressLine1, p.AddressLine2,
		p.City, p.State, p.Country, p.PostalCode, p.UpdatedAt)
	return err
}

func (r *Postgres) Get(ctx context.Context, userID string) (domain.Profile, error) {
	var p domain.Profile
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, date_of_birth, gender, address_line1, address_line2, city, state, country, postal_code, updated_at
		 FROM profile_profiles WHERE user_id = $1`, userID).
		Scan(&p.UserID, &p.DateOfBirth, &p.Gender, &p.AddressLine1, &p.AddressLine2,
			&p.City, &p.State, &p.Country, &p.PostalCode, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, apierror.ErrNotFound("profile not found")
	}
	return p, err
}
