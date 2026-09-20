// Package domain holds kyc-service business entities.
package domain

import "time"

// Status is the state of a KYC verification.
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusVerified Status = "VERIFIED"
	StatusRejected Status = "REJECTED"
)

// Verification is a single KYC submission and its review outcome.
type Verification struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	DocumentType string     `json:"document_type"` // e.g. PASSPORT, NATIONAL_ID
	DocumentID   string     `json:"document_id"`   // reference into document-service
	Status       Status     `json:"status"`
	Reason       string     `json:"reason,omitempty"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}
