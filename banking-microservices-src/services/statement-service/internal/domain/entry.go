// Package domain holds statement-service entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Direction classifies whether an entry increases or decreases an account.
const (
	DirectionDebit  = "DEBIT"
	DirectionCredit = "CREDIT"
)

// Entry is a single line in an account statement, derived from a transaction
// event. A transfer produces two entries (a DEBIT on the source account and a
// CREDIT on the destination); a deposit or withdrawal produces one.
type Entry struct {
	ID            string          `json:"id"`
	AccountID     string          `json:"account_id"`
	TransactionID string          `json:"transaction_id"`
	Type          string          `json:"type"`
	Direction     string          `json:"direction"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	CreatedAt     time.Time       `json:"created_at"`
}
