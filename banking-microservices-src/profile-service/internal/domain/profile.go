// Package domain holds profile-service business entities.
package domain

import "time"

// Profile holds a customer's extended personal details and primary address.
type Profile struct {
	UserID       string    `json:"user_id"`
	DateOfBirth  string    `json:"date_of_birth"` // ISO-8601 date
	Gender       string    `json:"gender"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 string    `json:"address_line2"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	PostalCode   string    `json:"postal_code"`
	UpdatedAt    time.Time `json:"updated_at"`
}
