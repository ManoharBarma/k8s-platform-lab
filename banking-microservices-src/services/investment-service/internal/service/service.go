// Package service implements investment-service use cases.
package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/investment-service/internal/domain"
	"banking-platform/services/investment-service/internal/repository"
)

// isNotFound reports whether err is a 404 apierror.
func isNotFound(err error) bool {
	var apiErr *apierror.Error
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

// Service holds investment-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Buy adds units of a symbol to an owner's holding, blending the average price.
func (s *Service) Buy(ctx context.Context, ownerID, symbol, unitsStr, priceStr string) (domain.Holding, error) {
	if symbol == "" {
		return domain.Holding{}, apierror.ErrBadRequest("symbol is required")
	}
	units, err := decimal.NewFromString(unitsStr)
	if err != nil || units.LessThanOrEqual(decimal.Zero) {
		return domain.Holding{}, apierror.ErrBadRequest("units must be a positive number")
	}
	price, err := decimal.NewFromString(priceStr)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return domain.Holding{}, apierror.ErrBadRequest("price must be a positive number")
	}

	now := time.Now().UTC()
	existing, err := s.repo.Get(ctx, ownerID, symbol)
	if err != nil {
		if !isNotFound(err) {
			return domain.Holding{}, err
		}
		h := domain.Holding{
			ID:        uuid.NewString(),
			OwnerID:   ownerID,
			Symbol:    symbol,
			Units:     units,
			AvgPrice:  price,
			UpdatedAt: now,
		}
		if err := s.repo.Upsert(ctx, h); err != nil {
			return domain.Holding{}, err
		}
		return h, nil
	}

	newUnits := existing.Units.Add(units)
	// new avg = (oldUnits*oldAvg + units*price) / (oldUnits+units)
	newAvg := existing.Units.Mul(existing.AvgPrice).Add(units.Mul(price)).Div(newUnits)
	existing.Units = newUnits
	existing.AvgPrice = newAvg
	existing.UpdatedAt = now
	if err := s.repo.Upsert(ctx, existing); err != nil {
		return domain.Holding{}, err
	}
	return existing, nil
}

// Sell reduces units of a symbol from an owner's holding, leaving avg price unchanged.
func (s *Service) Sell(ctx context.Context, ownerID, symbol, unitsStr string) (domain.Holding, error) {
	if symbol == "" {
		return domain.Holding{}, apierror.ErrBadRequest("symbol is required")
	}
	units, err := decimal.NewFromString(unitsStr)
	if err != nil || units.LessThanOrEqual(decimal.Zero) {
		return domain.Holding{}, apierror.ErrBadRequest("units must be a positive number")
	}

	h, err := s.repo.Get(ctx, ownerID, symbol)
	if err != nil {
		return domain.Holding{}, err
	}
	if units.GreaterThan(h.Units) {
		return domain.Holding{}, apierror.New(422, "insufficient_units", "not enough units")
	}

	h.Units = h.Units.Sub(units)
	h.UpdatedAt = time.Now().UTC()
	if h.Units.IsZero() {
		if err := s.repo.Delete(ctx, ownerID, symbol); err != nil {
			return domain.Holding{}, err
		}
		h.Units = decimal.Zero
		h.AvgPrice = decimal.Zero
		return h, nil
	}
	if err := s.repo.Upsert(ctx, h); err != nil {
		return domain.Holding{}, err
	}
	return h, nil
}

// Portfolio returns an owner's holdings.
func (s *Service) Portfolio(ctx context.Context, ownerID string) ([]domain.Holding, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}
