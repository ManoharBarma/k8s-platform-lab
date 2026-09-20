// Package service implements card-service use cases.
package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/card-service/internal/domain"
	"banking-platform/services/card-service/internal/repository"
)

// Service holds card-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Issue creates a new card for an owner, linked to an account.
func (s *Service) Issue(ctx context.Context, ownerID, accountID, cardType string) (domain.Card, error) {
	if cardType != "DEBIT" && cardType != "CREDIT" {
		return domain.Card{}, apierror.ErrBadRequest("type must be DEBIT or CREDIT")
	}
	if accountID == "" {
		return domain.Card{}, apierror.ErrBadRequest("account_id is required")
	}
	now := time.Now().UTC()
	card := domain.Card{
		ID:           uuid.NewString(),
		OwnerID:      ownerID,
		AccountID:    accountID,
		Type:         cardType,
		Network:      "VISA",
		MaskedNumber: maskPAN(generatePAN()),
		Status:       "ACTIVE",
		ExpiryMonth:  int(now.Month()),
		ExpiryYear:   now.Year() + 4,
		CreatedAt:    now,
	}
	if err := s.repo.Create(ctx, card); err != nil {
		return domain.Card{}, err
	}
	return card, nil
}

// ListByOwner returns an owner's cards.
func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]domain.Card, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// SetStatus updates a card's status, enforcing ownership.
func (s *Service) SetStatus(ctx context.Context, ownerID, id, status string) (domain.Card, error) {
	card, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Card{}, err
	}
	if card.OwnerID != ownerID {
		return domain.Card{}, apierror.ErrForbidden("not your card")
	}
	card.Status = status
	if err := s.repo.Update(ctx, card); err != nil {
		return domain.Card{}, err
	}
	return card, nil
}

// generatePAN returns a random 16-digit primary account number.
func generatePAN() string {
	const digits = 16
	buf := make([]byte, digits)
	for i := range buf {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		buf[i] = byte('0' + n.Int64())
	}
	return string(buf)
}

// maskPAN keeps only the last 4 digits of a PAN, discarding the rest.
func maskPAN(pan string) string {
	last4 := pan[len(pan)-4:]
	return "**** **** **** " + last4
}
