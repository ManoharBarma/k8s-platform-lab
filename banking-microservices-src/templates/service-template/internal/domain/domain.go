// Package domain contains the service's core business entities. Replace the
// sample Resource aggregate with the real entities for your service.
package domain

import "time"

// Resource is the sample aggregate managed by the template service.
type Resource struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
