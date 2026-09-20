// Package domain holds notification-service entities.
package domain

import "time"

// Notification is a message delivered to a recipient in response to a domain
// event consumed from Kafka.
type Notification struct {
	ID        string    `json:"id"`
	Recipient string    `json:"recipient"`
	Topic     string    `json:"topic"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
