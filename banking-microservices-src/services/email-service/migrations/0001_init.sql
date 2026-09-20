-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS email_messages (
    id         UUID PRIMARY KEY,
    to_addr    TEXT NOT NULL,
    subject    TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS email_created_at_idx ON email_messages (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_messages;
-- +goose StatementEnd
