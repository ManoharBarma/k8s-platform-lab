// Package domain holds sms-service business entities.
package domain

import "time"

// Message is a mock SMS dispatched to a phone number.
type Message struct {
	ID        string    `json:"id"`
	To        string    `json:"to"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
