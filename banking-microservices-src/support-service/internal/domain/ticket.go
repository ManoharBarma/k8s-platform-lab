// Package domain holds support-service business entities.
package domain

import "time"

// Ticket is a support request raised by a customer.
type Ticket struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"` // OPEN / IN_PROGRESS / RESOLVED / CLOSED
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
