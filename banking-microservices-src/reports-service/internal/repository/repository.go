// Package repository provides persistence for reports daily aggregates.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/services/reports-service/internal/domain"
)

// Repository stores daily transaction aggregates.
type Repository interface {
	// Add upserts the (date, currency) bucket: increments the count by 1 and
	// adds amount to the running total.
	Add(ctx context.Context, date, currency string, amount decimal.Decimal) error
	// List returns aggregates; if date == "" all rows ordered by date desc,
	// otherwise only rows for the given date.
	List(ctx context.Context, date string) ([]domain.DailyTotal, error)
}

// --- In-memory ---

type dailyKey struct {
	date     string
	currency string
}

// InMemory is a concurrency-safe in-memory aggregate store.
type InMemory struct {
	mu      sync.RWMutex
	buckets map[dailyKey]domain.DailyTotal
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{buckets: map[dailyKey]domain.DailyTotal{}}
}

func (r *InMemory) Add(_ context.Context, date, currency string, amount decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := dailyKey{date: date, currency: currency}
	t, ok := r.buckets[k]
	if !ok {
		t = domain.DailyTotal{Date: date, Currency: currency, TotalAmount: decimal.Zero}
	}
	t.Count++
	t.TotalAmount = t.TotalAmount.Add(amount)
	r.buckets[k] = t
	return nil
}

func (r *InMemory) List(_ context.Context, date string) ([]domain.DailyTotal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.DailyTotal
	for _, t := range r.buckets {
		if date == "" || t.Date == date {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date > out[j].Date })
	return out, nil
}

// --- Postgres ---

// Postgres persists daily aggregates under the reports_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Add(ctx context.Context, date, currency string, amount decimal.Decimal) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reports_daily (date, currency, count, total_amount)
		 VALUES ($1, $2, 1, $3::numeric)
		 ON CONFLICT (date, currency)
		 DO UPDATE SET count = reports_daily.count + 1,
		               total_amount = reports_daily.total_amount + EXCLUDED.total_amount`,
		date, currency, amount.String())
	return err
}

func (r *Postgres) List(ctx context.Context, date string) ([]domain.DailyTotal, error) {
	var rows interface {
		Next() bool
		Scan(...any) error
		Err() error
		Close()
	}
	var err error
	if date == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT date, currency, count, total_amount::text FROM reports_daily ORDER BY date DESC`)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT date, currency, count, total_amount::text FROM reports_daily WHERE date=$1 ORDER BY date DESC`, date)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DailyTotal
	for rows.Next() {
		var t domain.DailyTotal
		var amt string
		if err := rows.Scan(&t.Date, &t.Currency, &t.Count, &amt); err != nil {
			return nil, err
		}
		t.TotalAmount, _ = decimal.NewFromString(amt)
		out = append(out, t)
	}
	return out, rows.Err()
}
