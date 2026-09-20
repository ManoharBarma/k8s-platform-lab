-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS fraud_alerts (
    id             UUID PRIMARY KEY,
    transaction_id TEXT NOT NULL DEFAULT '',
    account_id     TEXT NOT NULL DEFAULT '',
    amount         NUMERIC(20,4) NOT NULL DEFAULT 0,
    currency       TEXT NOT NULL DEFAULT '',
    reason         TEXT NOT NULL DEFAULT '',
    score          INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS fraud_alerts_account_idx ON fraud_alerts (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS fraud_alerts_created_idx ON fraud_alerts (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fraud_alerts;
-- +goose StatementEnd
