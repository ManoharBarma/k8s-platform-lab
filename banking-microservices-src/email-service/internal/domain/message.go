// Package domain holds email-service business entities.
package domain

import "time"

// Message is a mock email that has been sent.
type Message struct {
	ID        string    `json:"id"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
