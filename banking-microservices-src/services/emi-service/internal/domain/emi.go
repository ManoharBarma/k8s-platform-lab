// Package domain holds emi-service business entities and amortization math.
package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Installment is one row of an amortization schedule.
type Installment struct {
	ID        string          `json:"id"`
	LoanID    string          `json:"loan_id"`
	Number    int             `json:"number"`
	DueDate   time.Time       `json:"due_date"`
	Principal decimal.Decimal `json:"principal"`
	Interest  decimal.Decimal `json:"interest"`
	Total     decimal.Decimal `json:"total"`
	Balance   decimal.Decimal `json:"balance"`
	Paid      bool            `json:"paid"`
	PaidAt    *time.Time      `json:"paid_at,omitempty"`
}

// GenerateSchedule builds a reducing-balance amortization schedule.
func GenerateSchedule(loanID string, principal decimal.Decimal, annualRatePct float64, tenureMonths int, start time.Time) []Installment {
	p, _ := principal.Float64()
	r := annualRatePct / 12 / 100

	var emi float64
	if r == 0 {
		emi = p / float64(tenureMonths)
	} else {
		pow := math.Pow(1+r, float64(tenureMonths))
		emi = p * r * pow / (pow - 1)
	}

	schedule := make([]Installment, 0, tenureMonths)
	balance := p
	for i := 1; i <= tenureMonths; i++ {
		interest := balance * r
		principalPart := emi - interest
		if i == tenureMonths {
			// Final installment absorbs rounding so the balance closes at zero.
			principalPart = balance
			emi = principalPart + interest
		}
		balance -= principalPart
		if balance < 0 {
			balance = 0
		}
		schedule = append(schedule, Installment{
			ID:        uuid.NewString(),
			LoanID:    loanID,
			Number:    i,
			DueDate:   start.AddDate(0, i, 0),
			Principal: decimal.NewFromFloat(principalPart).Round(2),
			Interest:  decimal.NewFromFloat(interest).Round(2),
			Total:     decimal.NewFromFloat(emi).Round(2),
			Balance:   decimal.NewFromFloat(balance).Round(2),
		})
	}
	return schedule
}
