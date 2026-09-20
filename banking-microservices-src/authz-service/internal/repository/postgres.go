package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/services/authz-service/internal/domain"
)

// Postgres persists the RBAC model in the shared database under the authz_
// prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) EnsureRole(ctx context.Context, name, description string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO authz_roles (name, description) VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description`,
		name, description)
	return err
}

func (r *Postgres) EnsurePermission(ctx context.Context, name, description string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO authz_permissions (name, description) VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description`,
		name, description)
	return err
}

func (r *Postgres) GrantPermission(ctx context.Context, role, permission string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO authz_role_permissions (role, permission) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, role, permission)
	return err
}

func (r *Postgres) AssignRole(ctx context.Context, subject, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO authz_user_roles (subject, role) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, subject, role)
	return err
}

func (r *Postgres) RolesForSubject(ctx context.Context, subject string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT role FROM authz_user_roles WHERE subject = $1`, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (r *Postgres) PermissionsForRoles(ctx context.Context, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT permission FROM authz_role_permissions WHERE role = ANY($1)`, roles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (r *Postgres) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `SELECT name, description FROM authz_roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.Name, &role.Description); err != nil {
			return nil, err
		}
		perms, err := r.PermissionsForRoles(ctx, []string{role.Name})
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
		out = append(out, role)
	}
	return out, rows.Err()
}

func scanStrings(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]string, error) {
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
