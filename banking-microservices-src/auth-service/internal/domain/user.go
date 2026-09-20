// Package domain holds auth-service business entities.
package domain

import "time"

// User is a platform login identity. It intentionally holds only
// authentication concerns; richer customer data lives in customer-service.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Roles        []string  `json:"roles"`
	CreatedAt    time.Time `json:"created_at"`
}
