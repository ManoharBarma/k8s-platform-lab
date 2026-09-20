// Package repository provides persistence for notifications.
package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"banking-platform/services/notification-service/internal/domain"
)

// Repository stores notifications.
type Repository interface {
	Append(ctx context.Context, n domain.Notification) error
	ListByRecipient(ctx context.Context, recipient string, limit int) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id string) error
}

// --- In-memory ---

// InMemory is a concurrency-safe in-memory notification store.
type InMemory struct {
	mu            sync.RWMutex
	notifications []domain.Notification
}

// NewInMemory builds an empty in-memory repository.
func NewInMemory() *InMemory { return &InMemory{} }

func (r *InMemory) Append(_ context.Context, n domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications = append(r.notifications, n)
	return nil
}

func (r *InMemory) ListByRecipient(_ context.Context, recipient string, limit int) ([]domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Notification
	for i := len(r.notifications) - 1; i >= 0; i-- {
		if r.notifications[i].Recipient == recipient {
			out = append(out, r.notifications[i])
			if len(out) >= limit {
				break
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemory) MarkRead(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.notifications {
		if r.notifications[i].ID == id {
			r.notifications[i].Read = true
			return nil
		}
	}
	return nil
}

// --- Postgres ---

// Postgres persists notifications under the notification_ prefix.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) Append(ctx context.Context, n domain.Notification) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notification_notifications (id, recipient, topic, message, read, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		n.ID, n.Recipient, n.Topic, n.Message, n.Read, n.CreatedAt)
	return err
}

func (r *Postgres) ListByRecipient(ctx context.Context, recipient string, limit int) ([]domain.Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, recipient, topic, message, read, created_at FROM notification_notifications WHERE recipient=$1 ORDER BY created_at DESC LIMIT $2`,
		recipient, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.Recipient, &n.Topic, &n.Message, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *Postgres) MarkRead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification_notifications SET read=true WHERE id=$1`, id)
	return err
}
