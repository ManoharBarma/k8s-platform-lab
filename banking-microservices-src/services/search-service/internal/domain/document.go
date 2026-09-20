// Package domain holds search-service entities.
package domain

import "time"

// Document is a searchable index entry derived from a consumed domain event.
// Content is a lowercase, denormalized text blob matched by search queries.
type Document struct {
	ID        string    `json:"id"`
	RefID     string    `json:"ref_id"`
	Kind      string    `json:"kind"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
