package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/templates/service-template/internal/domain"
)

// Postgres is a pgx-backed Repository. Note the tmpl_ table prefix: because the
// platform uses a single shared database, every service prefixes its tables to
// keep the schema organized and collision-free.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (r *Postgres) Create(ctx context.Context, res domain.Resource) (domain.Resource, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tmpl_resources (id, name, created_at) VALUES ($1, $2, $3)`,
		res.ID, res.Name, res.CreatedAt)
	return res, err
}

func (r *Postgres) Get(ctx context.Context, id string) (domain.Resource, error) {
	var res domain.Resource
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM tmpl_resources WHERE id = $1`, id).
		Scan(&res.ID, &res.Name, &res.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Resource{}, apierror.ErrNotFound("resource not found")
	}
	return res, err
}

func (r *Postgres) List(ctx context.Context) ([]domain.Resource, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, created_at FROM tmpl_resources ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(&res.ID, &res.Name, &res.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, rows.Err()
}
