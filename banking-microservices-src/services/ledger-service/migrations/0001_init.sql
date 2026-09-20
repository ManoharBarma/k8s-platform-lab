-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ledger_accounts (
    account_id UUID PRIMARY KEY,
    currency   TEXT NOT NULL,
    balance    NUMERIC(38,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ledger_transactions (
    id              UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL,
    currency        TEXT NOT NULL,
    reference       TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_transactions_idempotency_key_key UNIQUE (idempotency_key)
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id             UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES ledger_transactions(id),
    account_id     UUID NOT NULL REFERENCES ledger_accounts(account_id),
    direction      TEXT NOT NULL,
    amount         NUMERIC(38,2) NOT NULL,
    currency       TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ledger_entries_account_idx ON ledger_entries (account_id);
CREATE INDEX IF NOT EXISTS ledger_entries_txn_idx ON ledger_entries (transaction_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;
DROP TABLE IF EXISTS ledger_accounts;
-- +goose StatementEnd
