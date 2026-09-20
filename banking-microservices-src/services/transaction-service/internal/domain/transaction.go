// Package domain holds transaction-service business entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Type is the kind of money movement.
type Type string

const (
	TypeDeposit    Type = "DEPOSIT"
	TypeWithdrawal Type = "WITHDRAWAL"
	TypeTransfer   Type = "TRANSFER"
)

// Status is the transaction outcome.
type Status string

const (
	StatusPosted Status = "POSTED"
	StatusFailed Status = "FAILED"
)

// Transaction is a recorded money movement, backed by a balanced ledger posting.
type Transaction struct {
	ID             string          `json:"id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Type           Type            `json:"type"`
	FromAccountID  string          `json:"from_account_id,omitempty"`
	ToAccountID    string          `json:"to_account_id,omitempty"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Status         Status          `json:"status"`
	Reference      string          `json:"reference,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}
