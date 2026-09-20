// Package domain holds the double-entry ledger model.
//
// Balance convention: customer/bank accounts are modeled as liability-style
// accounts where a CREDIT increases the balance and a DEBIT decreases it
// (i.e. balance = ΣCREDIT − ΣDEBIT). A deposit credits the customer account; a
// withdrawal debits it; a transfer debits the source and credits the
// destination. Every posted transaction MUST balance: ΣDEBIT == ΣCREDIT.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Direction is the side of a ledger entry.
type Direction string

const (
	Debit  Direction = "DEBIT"
	Credit Direction = "CREDIT"
)

// Valid reports whether d is a known direction.
func (d Direction) Valid() bool { return d == Debit || d == Credit }

// Account is a ledger account holding a running balance in one currency.
type Account struct {
	AccountID string          `json:"account_id"`
	Currency  string          `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
}

// Entry is one leg of a double-entry transaction.
type Entry struct {
	AccountID string          `json:"account_id"`
	Direction Direction       `json:"direction"`
	Amount    decimal.Decimal `json:"amount"`
}

// Transaction is a balanced set of entries applied atomically.
type Transaction struct {
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Currency       string    `json:"currency"`
	Reference      string    `json:"reference"`
	Entries        []Entry   `json:"entries"`
	CreatedAt      time.Time `json:"created_at"`
}
