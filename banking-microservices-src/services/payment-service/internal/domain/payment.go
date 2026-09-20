// Package domain holds payment-service business entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Status is the payment outcome.
type Status string

const (
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Payment is an outbound payment from a customer account to a saved beneficiary.
// It debits the payer account and credits a system clearing account.
type Payment struct {
	ID             string          `json:"id"`
	PayerAccountID string          `json:"payer_account_id"`
	BeneficiaryID  string          `json:"beneficiary_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Status         Status          `json:"status"`
	Reference      string          `json:"reference"`
	IdempotencyKey string          `json:"idempotency_key"`
	CreatedAt      time.Time       `json:"created_at"`
}
