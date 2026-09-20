package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/pkg/apierror"
	"banking-platform/services/customer-service/internal/domain"
)

// Postgres persists customers in the shared database under the customer_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Create(ctx context.Context, c domain.Customer) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO customer_customers
		 (id, user_id, email, first_name, last_name, phone, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.UserID, c.Email, c.FirstName, c.LastName, c.Phone, c.Status, c.CreatedAt, c.UpdatedAt)
	if err != nil && strings.Contains(err.Error(), "customer_customers_user_id_key") {
		return apierror.ErrConflict("customer already exists for user")
	}
	return err
}

func (r *Postgres) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	return r.scanOne(ctx, `SELECT `+cols+` FROM customer_customers WHERE id = $1`, id)
}

func (r *Postgres) GetByUserID(ctx context.Context, userID string) (domain.Customer, error) {
	return r.scanOne(ctx, `SELECT `+cols+` FROM customer_customers WHERE user_id = $1`, userID)
}

func (r *Postgres) Update(ctx context.Context, c domain.Customer) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE customer_customers
		 SET email=$2, first_name=$3, last_name=$4, phone=$5, status=$6, updated_at=$7
		 WHERE id=$1`,
		c.ID, c.Email, c.FirstName, c.LastName, c.Phone, c.Status, c.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apierror.ErrNotFound("customer not found")
	}
	return nil
}

func (r *Postgres) List(ctx context.Context, limit, offset int) ([]domain.Customer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM customer_customers ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Customer
	for rows.Next() {
		c, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

const cols = "id, user_id, email, first_name, last_name, phone, status, created_at, updated_at"

type scanner interface{ Scan(...any) error }

func scanRow(s scanner) (domain.Customer, error) {
	var c domain.Customer
	err := s.Scan(&c.ID, &c.UserID, &c.Email, &c.FirstName, &c.LastName, &c.Phone, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *Postgres) scanOne(ctx context.Context, query, arg string) (domain.Customer, error) {
	c, err := scanRow(r.pool.QueryRow(ctx, query, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Customer{}, apierror.ErrNotFound("customer not found")
	}
	return c, err
}
