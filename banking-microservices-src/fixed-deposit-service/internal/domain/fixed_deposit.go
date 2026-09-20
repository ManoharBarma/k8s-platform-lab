// Package domain holds fixed-deposit-service business entities and financial math.
package domain

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// FixedDeposit is a customer's fixed deposit.
type FixedDeposit struct {
	ID             string          `json:"id"`
	OwnerID        string          `json:"owner_id"` // the JWT subject
	Principal      decimal.Decimal `json:"principal"`
	Currency       string          `json:"currency"`
	AnnualRatePct  float64         `json:"annual_rate_pct"`
	TenureMonths   int             `json:"tenure_months"`
	MaturityAmount decimal.Decimal `json:"maturity_amount"`
	Status         string          `json:"status"` // ACTIVE / MATURED / CLOSED
	CreatedAt      time.Time       `json:"created_at"`
	MaturesAt      time.Time       `json:"matures_at"`
}

// CalculateMaturity returns the maturity amount using compound-annual interest:
// maturity = principal · (1 + rate/100)^(tenureMonths/12).
func CalculateMaturity(principal decimal.Decimal, annualRatePct float64, tenureMonths int) decimal.Decimal {
	p, _ := principal.Float64()
	years := float64(tenureMonths) / 12
	maturity := p * math.Pow(1+annualRatePct/100, years)
	return decimal.NewFromFloat(maturity).Round(2)
}
