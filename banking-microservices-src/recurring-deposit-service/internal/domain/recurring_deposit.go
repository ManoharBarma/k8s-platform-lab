// Package domain holds recurring-deposit-service business entities and financial math.
package domain

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// RecurringDeposit is a customer's recurring deposit plan.
type RecurringDeposit struct {
	ID               string          `json:"id"`
	OwnerID          string          `json:"owner_id"`
	MonthlyAmount    decimal.Decimal `json:"monthly_amount"`
	Currency         string          `json:"currency"`
	AnnualRatePct    float64         `json:"annual_rate_pct"`
	TenureMonths     int             `json:"tenure_months"`
	InstallmentsPaid int             `json:"installments_paid"`
	MaturityAmount   decimal.Decimal `json:"maturity_amount"`
	Status           string          `json:"status"` // ACTIVE / MATURED / CLOSED
	CreatedAt        time.Time       `json:"created_at"`
}

// CalculateMaturity returns the maturity value of a recurring deposit using the
// standard RD formula with monthly compounding: M = P · (((1+i)^n − 1)/i) · (1+i),
// where i is the monthly rate and n the number of installments. A zero rate falls
// back to the plain sum of installments.
func CalculateMaturity(monthly decimal.Decimal, annualRatePct float64, tenureMonths int) decimal.Decimal {
	if tenureMonths <= 0 {
		return decimal.Zero
	}
	p, _ := monthly.Float64()
	if annualRatePct == 0 {
		return decimal.NewFromFloat(p * float64(tenureMonths)).Round(2)
	}
	i := annualRatePct / 12 / 100
	n := float64(tenureMonths)
	m := p * ((math.Pow(1+i, n) - 1) / i) * (1 + i)
	return decimal.NewFromFloat(m).Round(2)
}
