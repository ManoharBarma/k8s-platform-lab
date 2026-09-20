// Package domain holds document-service business entities.
package domain

import "time"

// Document is metadata about an object stored in S3. The object bytes live in
// S3 under Key; this record tracks ownership and type.
type Document struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Type        string    `json:"type"` // e.g. KYC_PASSPORT, STATEMENT
	Key         string    `json:"key"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}
