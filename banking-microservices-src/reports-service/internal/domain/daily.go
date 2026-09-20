// Package domain holds reports-service entities.
package domain

import "github.com/shopspring/decimal"

// DailyTotal is a per-day, per-currency aggregate of consumed transactions.
type DailyTotal struct {
	Date        string          `json:"date"`
	Currency    string          `json:"currency"`
	Count       int64           `json:"count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}
