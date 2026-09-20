-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sms_messages (
    id         UUID PRIMARY KEY,
    to_number  TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sms_messages_created_at_idx ON sms_messages (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sms_messages;
-- +goose StatementEnd
