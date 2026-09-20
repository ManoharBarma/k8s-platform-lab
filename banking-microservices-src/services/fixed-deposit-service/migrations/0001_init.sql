-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS fd_deposits (
    id               UUID PRIMARY KEY,
    owner_id         TEXT NOT NULL,
    principal        NUMERIC(38,2) NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'USD',
    annual_rate_pct  DOUBLE PRECISION NOT NULL DEFAULT 0,
    tenure_months    INTEGER NOT NULL,
    maturity_amount  NUMERIC(38,2) NOT NULL,
    status           TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    matures_at       TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS fd_owner_idx ON fd_deposits (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fd_deposits;
-- +goose StatementEnd
