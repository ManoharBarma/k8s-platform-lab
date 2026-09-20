-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payment_payments (
    id               UUID PRIMARY KEY,
    payer_account_id UUID NOT NULL,
    beneficiary_id   TEXT NOT NULL,
    amount           NUMERIC(38,2) NOT NULL,
    currency         TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'COMPLETED',
    reference        TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payment_payments_idempotency_key_key UNIQUE (idempotency_key)
);
CREATE INDEX IF NOT EXISTS payment_payer_idx ON payment_payments (payer_account_id);

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
DROP TABLE IF EXISTS payment_payments;
-- +goose StatementEnd
