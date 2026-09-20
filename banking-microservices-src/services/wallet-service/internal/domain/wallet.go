// Package domain holds wallet-service business entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TxnType is the kind of wallet movement.
type TxnType string

const (
	TxnTopUp TxnType = "TOPUP"
	TxnSpend TxnType = "SPEND"
)

// Wallet is a customer's stored-value balance.
type Wallet struct {
	UserID    string          `json:"user_id"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Transaction is a wallet top-up or spend.
type Transaction struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Type           TxnType         `json:"type"`
	Amount         decimal.Decimal `json:"amount"`
	IdempotencyKey string          `json:"idempotency_key"`
	CreatedAt      time.Time       `json:"created_at"`
}
