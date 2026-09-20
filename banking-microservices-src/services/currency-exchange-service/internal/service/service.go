// Package service implements currency-exchange-service use cases.
package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/currency-exchange-service/internal/domain"
	"banking-platform/services/currency-exchange-service/internal/repository"
)

// isNotFound reports whether err is an apierror with a 404 status.
func isNotFound(err error) bool {
	var ae *apierror.Error
	return errors.As(err, &ae) && ae.Status == http.StatusNotFound
}

// Service holds currency-exchange-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Seed installs the platform's default exchange rates (units per 1 USD). It is
// idempotent.
func (s *Service) Seed(ctx context.Context) error {
	rates := map[string]string{
		"USD": "1",
		"EUR": "0.92",
		"GBP": "0.79",
		"INR": "83.0",
		"JPY": "149.0",
		"AUD": "1.52",
		"CAD": "1.36",
	}
	for code, perUSD := range rates {
		v, err := decimal.NewFromString(perUSD)
		if err != nil {
			return err
		}
		if err := s.repo.Upsert(ctx, code, v); err != nil {
			return err
		}
	}
	s.log.Info("seeded fx rates")
	return nil
}

// List returns all exchange rates ordered by code.
func (s *Service) List(ctx context.Context) ([]domain.Rate, error) {
	return s.repo.List(ctx)
}

// Convert converts amountStr from one currency to another using USD as the base.
// It returns the converted result (rounded to 2 dp) and the effective rate
// (to per from, rounded to 6 dp).
func (s *Service) Convert(ctx context.Context, from, to, amountStr string) (result decimal.Decimal, rate decimal.Decimal, err error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil || !amount.IsPositive() {
		return decimal.Decimal{}, decimal.Decimal{}, apierror.ErrBadRequest("amount must be a positive number")
	}

	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	rFrom, err := s.repo.Get(ctx, from)
	if err != nil {
		if isNotFound(err) {
			return decimal.Decimal{}, decimal.Decimal{}, apierror.ErrBadRequest("unsupported currency")
		}
		return decimal.Decimal{}, decimal.Decimal{}, err
	}
	rTo, err := s.repo.Get(ctx, to)
	if err != nil {
		if isNotFound(err) {
			return decimal.Decimal{}, decimal.Decimal{}, apierror.ErrBadRequest("unsupported currency")
		}
		return decimal.Decimal{}, decimal.Decimal{}, err
	}

	usd := amount.Div(rFrom)
	result = usd.Mul(rTo).Round(2)
	rate = rTo.Div(rFrom).Round(6)
	return result, rate, nil
}
