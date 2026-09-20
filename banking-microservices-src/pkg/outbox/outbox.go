// Package outbox implements the transactional outbox pattern: events are
// written to an outbox table in the SAME database transaction as the business
// state change, then a background relay reliably publishes them to Kafka. This
// avoids the dual-write problem (DB commit succeeds but Kafka publish fails, or
// vice-versa) that would otherwise corrupt an event-driven banking system.
package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DDL creates the shared outbox table. Services include this in their
// migrations (it is idempotent). A single table is used because the platform
// shares one database; each service filters by `source`.
const DDL = `
CREATE TABLE IF NOT EXISTS outbox_events (
    id         UUID PRIMARY KEY,
    source     TEXT NOT NULL,
    topic      TEXT NOT NULL,
    key        TEXT NOT NULL DEFAULT '',
    payload    BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS outbox_events_unsent_idx
    ON outbox_events (source, created_at) WHERE sent_at IS NULL;
`

// Event is a message queued for reliable publication.
type Event struct {
	ID      string
	Source  string
	Topic   string
	Key     string
	Payload []byte
}

// Insert writes an event within an existing transaction so it commits atomically
// with the business write.
func Insert(ctx context.Context, tx pgx.Tx, e Event) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO outbox_events (id, source, topic, key, payload) VALUES ($1,$2,$3,$4,$5)`,
		e.ID, e.Source, e.Topic, e.Key, e.Payload)
	return err
}

// Publisher publishes a message to a topic (implemented by pkg/kafka.Producer).
type Publisher interface {
	Publish(ctx context.Context, topic, key string, value []byte) error
}

// Relay polls the outbox for this service's unsent events and publishes them.
type Relay struct {
	pool     *pgxpool.Pool
	pub      Publisher
	source   string
	interval time.Duration
	batch    int
	log      *zap.Logger
}

// NewRelay builds a relay for the given source (service name).
func NewRelay(pool *pgxpool.Pool, pub Publisher, source string, log *zap.Logger) *Relay {
	return &Relay{pool: pool, pub: pub, source: source, interval: 2 * time.Second, batch: 100, log: log}
}

// Run polls until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.flush(ctx); err != nil {
				r.log.Warn("outbox flush failed", zap.Error(err))
			}
		}
	}
}

func (r *Relay) flush(ctx context.Context) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, topic, key, payload FROM outbox_events
		 WHERE source = $1 AND sent_at IS NULL
		 ORDER BY created_at LIMIT $2`, r.source, r.batch)
	if err != nil {
		return err
	}
	type row struct {
		id, topic, key string
		payload        []byte
	}
	var pending []row
	for rows.Next() {
		var x row
		if err := rows.Scan(&x.id, &x.topic, &x.key, &x.payload); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, x)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, x := range pending {
		if err := r.pub.Publish(ctx, x.topic, x.key, x.payload); err != nil {
			return err // stop; retry on next tick to preserve ordering
		}
		if _, err := r.pool.Exec(ctx, `UPDATE outbox_events SET sent_at = now() WHERE id = $1`, x.id); err != nil {
			return err
		}
	}
	if len(pending) > 0 {
		r.log.Debug("outbox published", zap.Int("count", len(pending)))
	}
	return nil
}
