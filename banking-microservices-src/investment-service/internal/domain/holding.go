// Package domain holds investment-service business entities.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Holding is an owner's position in a single symbol, unique per (owner, symbol).
type Holding struct {
	ID        string          `json:"id"`
	OwnerID   string          `json:"owner_id"`
	Symbol    string          `json:"symbol"`
	Units     decimal.Decimal `json:"units"`
	AvgPrice  decimal.Decimal `json:"avg_price"`
	UpdatedAt time.Time       `json:"updated_at"`
}
