-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS beneficiary_beneficiaries (
    id             UUID PRIMARY KEY,
    owner_id       TEXT NOT NULL,
    name           TEXT NOT NULL,
    account_number TEXT NOT NULL,
    bank_name      TEXT NOT NULL DEFAULT '',
    routing_code   TEXT NOT NULL DEFAULT '',
    currency       TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS beneficiary_owner_idx ON beneficiary_beneficiaries (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS beneficiary_beneficiaries;
-- +goose StatementEnd
