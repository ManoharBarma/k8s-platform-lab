// Package service implements statement-service: it consumes transaction events
// from Kafka and derives per-account statement entries, and serves queries.
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/statement-service/internal/domain"
	"banking-platform/services/statement-service/internal/repository"
)

// Service holds statement-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// transactionEvent is the transaction payload consumed from Kafka.
type transactionEvent struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	FromAccountID string          `json:"from_account_id"`
	ToAccountID   string          `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	CreatedAt     time.Time       `json:"created_at"`
}

// Handle is the Kafka consumer callback: it parses a transaction event and
// records the resulting statement entries (a DEBIT for the source account and a
// CREDIT for the destination account, whichever are present).
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	var ev transactionEvent
	if err := json.Unmarshal(m.Value, &ev); err != nil {
		s.log.Warn("skipping unparseable transaction event", zap.Error(err))
		return nil
	}

	var entries []domain.Entry
	if ev.FromAccountID != "" {
		entries = append(entries, s.newEntry(ev, ev.FromAccountID, domain.DirectionDebit))
	}
	if ev.ToAccountID != "" {
		entries = append(entries, s.newEntry(ev, ev.ToAccountID, domain.DirectionCredit))
	}
	if len(entries) == 0 {
		return nil
	}
	return s.repo.CreateMany(ctx, entries)
}

func (s *Service) newEntry(ev transactionEvent, accountID, direction string) domain.Entry {
	created := ev.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	return domain.Entry{
		ID:            uuid.NewString(),
		AccountID:     accountID,
		TransactionID: ev.ID,
		Type:          ev.Type,
		Direction:     direction,
		Amount:        ev.Amount,
		Currency:      ev.Currency,
		CreatedAt:     created,
	}
}

// ListByAccount returns recent statement entries for an account, newest first.
func (s *Service) ListByAccount(ctx context.Context, accountID string, limit int) ([]domain.Entry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.ListByAccount(ctx, accountID, limit)
}
