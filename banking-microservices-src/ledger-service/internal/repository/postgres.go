package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"banking-platform/pkg/apierror"
	"banking-platform/services/ledger-service/internal/domain"
)

// Postgres is the authoritative double-entry ledger. Postings are applied inside
// a single serializable-safe transaction: balances and journal entries commit
// together or not at all.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres builds a Postgres-backed ledger.
func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (r *Postgres) OpenAccount(ctx context.Context, accountID, currency string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ledger_accounts (account_id, currency, balance) VALUES ($1, $2, 0)
		 ON CONFLICT (account_id) DO NOTHING`, accountID, currency)
	return err
}

func (r *Postgres) Post(ctx context.Context, t domain.Transaction) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Idempotency: insert the transaction header; a pre-existing key means the
	// posting was already applied.
	_, err = tx.Exec(ctx,
		`INSERT INTO ledger_transactions (id, idempotency_key, currency, reference)
		 VALUES ($1,$2,$3,$4) ON CONFLICT (idempotency_key) DO NOTHING`,
		t.ID, t.IdempotencyKey, t.Currency, t.Reference)
	if err != nil {
		return false, err
	}
	var storedID string
	if err := tx.QueryRow(ctx,
		`SELECT id FROM ledger_transactions WHERE idempotency_key = $1`, t.IdempotencyKey).
		Scan(&storedID); err != nil {
		return false, err
	}
	if storedID != t.ID {
		// Key already applied by a previous transaction — no-op.
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	for _, e := range t.Entries {
		delta := e.Amount
		if e.Direction == domain.Debit {
			delta = delta.Neg()
		}
		ct, err := tx.Exec(ctx,
			`UPDATE ledger_accounts SET balance = balance + $2::numeric WHERE account_id = $1`,
			e.AccountID, delta.String())
		if err != nil {
			return false, err
		}
		if ct.RowsAffected() == 0 {
			return false, apierror.ErrBadRequest("ledger account not found: " + e.AccountID)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO ledger_entries (id, transaction_id, account_id, direction, amount, currency)
			 VALUES ($1,$2,$3,$4,$5::numeric,$6)`,
			uuid.NewString(), t.ID, e.AccountID, string(e.Direction), e.Amount.String(), t.Currency); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return false, nil
}

func (r *Postgres) Balance(ctx context.Context, accountID string) (decimal.Decimal, string, error) {
	var balanceStr, currency string
	err := r.pool.QueryRow(ctx,
		`SELECT balance::text, currency FROM ledger_accounts WHERE account_id = $1`, accountID).
		Scan(&balanceStr, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, "", apierror.ErrNotFound("ledger account not found")
	}
	if err != nil {
		return decimal.Zero, "", err
	}
	bal, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return decimal.Zero, "", err
	}
	return bal, currency, nil
}
