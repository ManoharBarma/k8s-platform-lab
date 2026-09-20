// Package money provides an exact decimal monetary type. Never represent money
// as a float — use Money everywhere amounts are handled.
package money

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Money is an exact monetary amount in an ISO-4217 currency.
type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// New builds Money from a decimal amount and currency.
func New(amount decimal.Decimal, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

// FromString parses a decimal amount string (e.g. "100.50").
func FromString(amount, currency string) (Money, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount %q: %w", amount, err)
	}
	return Money{Amount: d, Currency: currency}, nil
}

// Zero returns a zero amount in the given currency.
func Zero(currency string) Money {
	return Money{Amount: decimal.Zero, Currency: currency}
}

// Add returns m + other, erroring on currency mismatch.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s != %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}

// Sub returns m - other, erroring on currency mismatch.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s != %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount.Sub(other.Amount), Currency: m.Currency}, nil
}

// IsNegative reports whether the amount is below zero.
func (m Money) IsNegative() bool { return m.Amount.IsNegative() }

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool { return m.Amount.IsZero() }

// GreaterThanOrEqual reports whether m >= other (currencies must match).
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if m.Currency != other.Currency {
		return false, fmt.Errorf("currency mismatch: %s != %s", m.Currency, other.Currency)
	}
	return m.Amount.GreaterThanOrEqual(other.Amount), nil
}

// String renders the amount with 2 decimal places and its currency.
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.StringFixed(2), m.Currency)
}
