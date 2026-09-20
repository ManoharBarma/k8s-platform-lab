-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS support_tickets (
    id         UUID PRIMARY KEY,
    owner_id   TEXT NOT NULL,
    subject    TEXT NOT NULL,
    body       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS support_owner_idx ON support_tickets (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS support_tickets;
-- +goose StatementEnd
