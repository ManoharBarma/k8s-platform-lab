// Package domain holds card-service business entities.
package domain

import "time"

// Card is a payment card owned by a customer and linked to an account.
type Card struct {
	ID           string    `json:"id"`
	OwnerID      string    `json:"owner_id"`
	AccountID    string    `json:"account_id"`
	Type         string    `json:"type"`          // DEBIT / CREDIT
	Network      string    `json:"network"`       // e.g. VISA
	MaskedNumber string    `json:"masked_number"` // **** **** **** 1234
	Status       string    `json:"status"`        // ACTIVE / BLOCKED
	ExpiryMonth  int       `json:"expiry_month"`
	ExpiryYear   int       `json:"expiry_year"`
	CreatedAt    time.Time `json:"created_at"`
}
