// Package domain holds beneficiary-service business entities.
package domain

import "time"

// Beneficiary is a saved payee owned by a customer.
type Beneficiary struct {
	ID            string    `json:"id"`
	OwnerID       string    `json:"owner_id"`
	Name          string    `json:"name"`
	AccountNumber string    `json:"account_number"`
	BankName      string    `json:"bank_name"`
	RoutingCode   string    `json:"routing_code"` // IFSC / SWIFT / routing number
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}
