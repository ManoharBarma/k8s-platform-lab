// Package domain holds customer-service business entities.
package domain

import "time"

// Status is the lifecycle state of a customer record.
type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusClosed    Status = "CLOSED"
)

// Customer is the master record for a banking customer. It is linked 1:1 to a
// login identity in auth-service via UserID (the JWT subject).
type Customer struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
