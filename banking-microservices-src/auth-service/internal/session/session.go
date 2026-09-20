// Package session stores active refresh tokens so they can be validated and
// revoked (logout). A Redis-backed store is used in production; an in-memory
// store is the fallback when Redis is disabled.
package session

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store tracks valid refresh tokens.
type Store interface {
	Save(ctx context.Context, token, subject string, ttl time.Duration) error
	Exists(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
}

const keyPrefix = "auth:refresh:"

// RedisStore is a Redis-backed session store.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore builds a Redis-backed store.
func NewRedisStore(client *redis.Client) *RedisStore { return &RedisStore{client: client} }

func (s *RedisStore) Save(ctx context.Context, token, subject string, ttl time.Duration) error {
	return s.client.Set(ctx, keyPrefix+token, subject, ttl).Err()
}

func (s *RedisStore) Exists(ctx context.Context, token string) (bool, error) {
	n, err := s.client.Exists(ctx, keyPrefix+token).Result()
	return n > 0, err
}

func (s *RedisStore) Delete(ctx context.Context, token string) error {
	return s.client.Del(ctx, keyPrefix+token).Err()
}

// MemoryStore is an in-memory session store with lazy expiry.
type MemoryStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time // token -> expiry
}

// NewMemoryStore builds an in-memory store.
func NewMemoryStore() *MemoryStore { return &MemoryStore{tokens: map[string]time.Time{}} }

func (s *MemoryStore) Save(_ context.Context, token, _ string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = time.Now().Add(ttl)
	return nil
}

func (s *MemoryStore) Exists(_ context.Context, token string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.tokens[token]
	if !ok {
		return false, nil
	}
	if time.Now().After(exp) {
		delete(s.tokens, token)
		return false, nil
	}
	return true, nil
}

func (s *MemoryStore) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
	return nil
}
