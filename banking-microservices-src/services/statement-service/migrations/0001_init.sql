-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS statement_entries (
    id             UUID PRIMARY KEY,
    account_id     TEXT NOT NULL,
    transaction_id TEXT NOT NULL DEFAULT '',
    type           TEXT NOT NULL DEFAULT '',
    direction      TEXT NOT NULL,
    amount         NUMERIC(38,2) NOT NULL,
    currency       TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS statement_entries_account_id_idx ON statement_entries (account_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS statement_entries;
-- +goose StatementEnd
