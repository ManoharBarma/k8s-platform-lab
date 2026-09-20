-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_notifications (
    id         UUID PRIMARY KEY,
    recipient  TEXT NOT NULL DEFAULT '',
    topic      TEXT NOT NULL,
    message    TEXT NOT NULL,
    read       BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS notification_notifications_recipient_idx ON notification_notifications (recipient, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_notifications;
-- +goose StatementEnd
