package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/auth-service/internal/domain"
)

// Postgres persists users in the shared database under the auth_ prefix. Roles
// are stored comma-separated for simplicity.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (r *Postgres) Create(ctx context.Context, u domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO auth_users (id, email, password_hash, full_name, roles, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, strings.ToLower(u.Email), u.PasswordHash, u.FullName,
		strings.Join(u.Roles, ","), u.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "auth_users_email_key") {
		return apierror.ErrConflict("email already registered")
	}
	return err
}

func (r *Postgres) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, password_hash, full_name, roles, created_at
		 FROM auth_users WHERE email = $1`, strings.ToLower(email))
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, password_hash, full_name, roles, created_at
		 FROM auth_users WHERE id = $1`, id)
}

func (r *Postgres) scanOne(ctx context.Context, query string, arg any) (domain.User, error) {
	var u domain.User
	var roles string
	err := r.pool.QueryRow(ctx, query, arg).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &roles, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, apierror.ErrNotFound("user not found")
	}
	if err != nil {
		return domain.User{}, err
	}
	if roles != "" {
		u.Roles = strings.Split(roles, ",")
	}
	return u, nil
}
