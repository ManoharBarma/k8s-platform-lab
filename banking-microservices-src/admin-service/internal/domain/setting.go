// Package domain holds admin-service business entities.
package domain

import "time"

// Setting is a platform-wide feature flag or configuration value that admins
// control and other services can read.
type Setting struct {
	Key       string    `json:"key"`   // unique setting identifier
	Value     string    `json:"value"` // setting value
	UpdatedAt time.Time `json:"updated_at"`
}
