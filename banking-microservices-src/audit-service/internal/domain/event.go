// Package domain holds audit-service entities.
package domain

import (
	"encoding/json"
	"time"
)

// Event is an immutable audit record capturing a domain event consumed from
// Kafka. The raw event payload is preserved verbatim.
type Event struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	Key       string          `json:"key"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}
