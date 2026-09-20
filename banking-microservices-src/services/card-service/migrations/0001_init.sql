-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS card_cards (
    id            UUID PRIMARY KEY,
    owner_id      TEXT NOT NULL,
    account_id    TEXT NOT NULL,
    type          TEXT NOT NULL,
    network       TEXT NOT NULL DEFAULT 'VISA',
    masked_number TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'ACTIVE',
    expiry_month  INTEGER NOT NULL,
    expiry_year   INTEGER NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS card_owner_idx ON card_cards (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS card_cards;
-- +goose StatementEnd
