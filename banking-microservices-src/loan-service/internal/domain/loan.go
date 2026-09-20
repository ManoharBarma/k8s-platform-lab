// Package domain holds loan-service business entities and financial math.
package domain

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// Status is the loan lifecycle state.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusApproved  Status = "APPROVED"
	StatusDisbursed Status = "DISBURSED"
	StatusRejected  Status = "REJECTED"
	StatusClosed    Status = "CLOSED"
)

// Loan is a customer loan.
type Loan struct {
	ID           string          `json:"id"`
	BorrowerID   string          `json:"borrower_id"`
	AccountID    string          `json:"account_id"` // account to disburse into
	Principal    decimal.Decimal `json:"principal"`
	Currency     string          `json:"currency"`
	AnnualRate   float64         `json:"annual_rate_pct"`
	TenureMonths int             `json:"tenure_months"`
	EMIAmount    decimal.Decimal `json:"emi_amount"`
	Outstanding  decimal.Decimal `json:"outstanding"`
	Status       Status          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	DisbursedAt  *time.Time      `json:"disbursed_at,omitempty"`
}

// CalculateEMI returns the equated monthly installment using the standard
// reducing-balance formula: EMI = P·r·(1+r)^n / ((1+r)^n − 1), where r is the
// monthly rate. A zero rate falls back to straight-line division.
func CalculateEMI(principal decimal.Decimal, annualRatePct float64, tenureMonths int) decimal.Decimal {
	if tenureMonths <= 0 {
		return decimal.Zero
	}
	p, _ := principal.Float64()
	if annualRatePct == 0 {
		return decimal.NewFromFloat(p / float64(tenureMonths)).Round(2)
	}
	r := annualRatePct / 12 / 100
	pow := math.Pow(1+r, float64(tenureMonths))
	emi := p * r * pow / (pow - 1)
	return decimal.NewFromFloat(emi).Round(2)
}
