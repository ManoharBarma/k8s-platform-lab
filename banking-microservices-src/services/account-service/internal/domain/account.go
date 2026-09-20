// Package domain holds account-service business entities.
package domain

import "time"

// Type is the kind of bank account.
type Type string

const (
	TypeSavings Type = "SAVINGS"
	TypeCurrent Type = "CURRENT"
)

// Status is the account lifecycle state.
type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusFrozen Status = "FROZEN"
	StatusClosed Status = "CLOSED"
)

// Account is a customer bank account. Its balance is NOT stored here — the
// ledger service is the source of truth and is queried on demand.
type Account struct {
	ID            string    `json:"id"`
	AccountNumber string    `json:"account_number"`
	CustomerID    string    `json:"customer_id"`
	Type          Type      `json:"type"`
	Currency      string    `json:"currency"`
	Status        Status    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
