-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rd_deposits (
    id                UUID PRIMARY KEY,
    owner_id          TEXT NOT NULL,
    monthly_amount    NUMERIC(20,2) NOT NULL,
    currency          TEXT NOT NULL DEFAULT '',
    annual_rate_pct   DOUBLE PRECISION NOT NULL DEFAULT 0,
    tenure_months     INTEGER NOT NULL DEFAULT 0,
    installments_paid INTEGER NOT NULL DEFAULT 0,
    maturity_amount   NUMERIC(20,2) NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS rd_owner_idx ON rd_deposits (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rd_deposits;
-- +goose StatementEnd
