// Package domain holds currency-exchange-service business entities.
package domain

import "github.com/shopspring/decimal"

// Rate is an exchange rate for a currency, expressed as units of this currency
// per 1 US dollar. The base currency is USD (PerUSD = 1).
type Rate struct {
	Code   string          `json:"code"`    // ISO currency code, e.g. "EUR"
	PerUSD decimal.Decimal `json:"per_usd"` // units of this currency per 1 USD
}
