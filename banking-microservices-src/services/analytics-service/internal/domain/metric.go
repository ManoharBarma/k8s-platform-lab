// Package domain holds analytics-service entities.
package domain

import "github.com/shopspring/decimal"

// Metric is a per-transaction-type aggregate. Key is the transaction type
// (e.g. "DEPOSIT", "WITHDRAWAL", "TRANSFER"); Count and Sum accumulate the
// number of events and total amount observed for that type.
type Metric struct {
	Key   string          `json:"key"`
	Count int64           `json:"count"`
	Sum   decimal.Decimal `json:"sum"`
}
