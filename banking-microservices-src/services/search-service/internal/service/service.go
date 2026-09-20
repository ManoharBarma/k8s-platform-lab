// Package service implements search-service: it consumes domain events, builds
// searchable index documents, and serves full-text queries over them.
package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/kafka"
	"banking-platform/services/search-service/internal/domain"
	"banking-platform/services/search-service/internal/repository"
)

// Service holds search-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// transactionEvent is the subset of a banking.transactions event indexed for
// search. Amount is kept as raw JSON so its text can be included verbatim.
type transactionEvent struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	FromAccountID string          `json:"from_account_id"`
	ToAccountID   string          `json:"to_account_id"`
	Currency      string          `json:"currency"`
	Reference     string          `json:"reference"`
	Amount        json.RawMessage `json:"amount"`
}

// Handle is the Kafka consumer callback: it parses the event, builds a
// lowercase searchable content blob, and stores an index document.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	var ev transactionEvent
	if err := json.Unmarshal(m.Value, &ev); err != nil {
		return err
	}

	parts := []string{ev.ID, ev.Type, ev.FromAccountID, ev.ToAccountID, ev.Currency, ev.Reference}
	if len(ev.Amount) > 0 {
		parts = append(parts, string(ev.Amount))
	}
	content := strings.ToLower(strings.Join(parts, " "))

	return s.repo.Index(ctx, domain.Document{
		ID:        uuid.NewString(),
		RefID:     ev.ID,
		Kind:      "transaction",
		Content:   content,
		CreatedAt: time.Now().UTC(),
	})
}

// Search returns index documents matching q, most recent first.
func (s *Service) Search(ctx context.Context, q string, limit int) ([]domain.Document, error) {
	if q == "" {
		return nil, apierror.ErrBadRequest("q is required")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.Search(ctx, q, limit)
}
