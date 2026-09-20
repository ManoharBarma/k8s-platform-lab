-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallet_wallets (
    user_id    TEXT PRIMARY KEY,
    balance    NUMERIC(38,2) NOT NULL DEFAULT 0,
    currency   TEXT NOT NULL DEFAULT 'USD',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id              UUID PRIMARY KEY,
    user_id         TEXT NOT NULL,
    type            TEXT NOT NULL,
    amount          NUMERIC(38,2) NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wallet_transactions_idempotency_key_key UNIQUE (idempotency_key)
);
CREATE INDEX IF NOT EXISTS wallet_transactions_user_idx ON wallet_transactions (user_id);

CREATE TABLE IF NOT EXISTS outbox_events (
    id         UUID PRIMARY KEY,
    source     TEXT NOT NULL,
    topic      TEXT NOT NULL,
    key        TEXT NOT NULL DEFAULT '',
    payload    BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS outbox_events_unsent_idx
    ON outbox_events (source, created_at) WHERE sent_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallet_wallets;
-- +goose StatementEnd
