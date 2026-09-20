// Package repository provides persistence for the ledger.
package repository

import (
	"context"

	"github.com/shopspring/decimal"

	"banking-platform/services/ledger-service/internal/domain"
)

// Repository stores ledger accounts and posts balanced transactions.
type Repository interface {
	// OpenAccount creates a ledger account (idempotent).
	OpenAccount(ctx context.Context, accountID, currency string) error
	// Post applies a balanced transaction atomically. It is idempotent on the
	// transaction's IdempotencyKey; when the key was already applied it returns
	// duplicate=true and makes no changes.
	Post(ctx context.Context, t domain.Transaction) (duplicate bool, err error)
	// Balance returns an account's current balance and currency.
	Balance(ctx context.Context, accountID string) (decimal.Decimal, string, error)
}
