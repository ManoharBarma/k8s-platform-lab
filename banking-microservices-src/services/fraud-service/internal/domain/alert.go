// Package domain holds fraud-service entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Alert is a fraud alert raised when a consumed domain event trips a fraud rule.
type Alert struct {
	ID            string          `json:"id"`
	TransactionID string          `json:"transaction_id"`
	AccountID     string          `json:"account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Reason        string          `json:"reason"`
	Score         int             `json:"score"`
	CreatedAt     time.Time       `json:"created_at"`
}
