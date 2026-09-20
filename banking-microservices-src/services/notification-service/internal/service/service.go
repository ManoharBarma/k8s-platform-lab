// Package service implements notification-service: it turns every consumed
// domain event into a notification for the affected recipient and serves
// queries over stored notifications.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/kafka"
	"banking-platform/services/notification-service/internal/domain"
	"banking-platform/services/notification-service/internal/repository"
)

// Service holds notification-service dependencies.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// eventPayload is a minimal view of a domain event payload, used to enrich the
// human-readable notification message when the fields are present.
type eventPayload struct {
	Type   string `json:"type"`
	Amount string `json:"amount"`
}

// Handle is the Kafka consumer callback: it builds a notification for the
// message recipient and stores it unread.
func (s *Service) Handle(ctx context.Context, m kafka.Message) error {
	message := fmt.Sprintf("Event on %s for %s", m.Topic, m.Key)
	var p eventPayload
	if err := json.Unmarshal(m.Value, &p); err == nil {
		if p.Type != "" && p.Amount != "" {
			message = fmt.Sprintf("Event on %s for %s: %s of %s", m.Topic, m.Key, p.Type, p.Amount)
		} else if p.Type != "" {
			message = fmt.Sprintf("Event on %s for %s: %s", m.Topic, m.Key, p.Type)
		}
	}
	return s.repo.Append(ctx, domain.Notification{
		ID:        uuid.NewString(),
		Recipient: m.Key,
		Topic:     m.Topic,
		Message:   message,
		Read:      false,
		CreatedAt: time.Now().UTC(),
	})
}

// ListByRecipient returns recent notifications for a recipient.
func (s *Service) ListByRecipient(ctx context.Context, recipient string, limit int) ([]domain.Notification, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.ListByRecipient(ctx, recipient, limit)
}

// MarkRead marks a notification as read.
func (s *Service) MarkRead(ctx context.Context, id string) error {
	return s.repo.MarkRead(ctx, id)
}
