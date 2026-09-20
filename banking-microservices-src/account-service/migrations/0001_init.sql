-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS account_accounts (
    id             UUID PRIMARY KEY,
    account_number TEXT NOT NULL,
    customer_id    TEXT NOT NULL,
    type           TEXT NOT NULL,
    currency       TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT account_accounts_number_key UNIQUE (account_number)
);
CREATE INDEX IF NOT EXISTS account_accounts_customer_idx ON account_accounts (customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS account_accounts;
-- +goose StatementEnd
