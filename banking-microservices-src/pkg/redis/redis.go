// Package redis provides a go-redis client with OpenTelemetry tracing enabled,
// so cache operations show up in distributed traces.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Config controls the Redis connection.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// New creates and verifies a Redis client with tracing instrumentation.
func New(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("instrument redis tracing: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return client, nil
}
