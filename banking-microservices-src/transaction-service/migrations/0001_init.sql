-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transaction_transactions (
    id              UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL,
    type            TEXT NOT NULL,
    from_account_id UUID,
    to_account_id   UUID,
    amount          NUMERIC(38,2) NOT NULL,
    currency        TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'POSTED',
    reference       TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT transaction_transactions_idempotency_key_key UNIQUE (idempotency_key)
);
CREATE INDEX IF NOT EXISTS transaction_from_idx ON transaction_transactions (from_account_id);
CREATE INDEX IF NOT EXISTS transaction_to_idx ON transaction_transactions (to_account_id);

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
DROP TABLE IF EXISTS transaction_transactions;
-- +goose StatementEnd
