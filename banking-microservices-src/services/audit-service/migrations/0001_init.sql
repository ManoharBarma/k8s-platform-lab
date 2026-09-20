-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS audit_events (
    id         UUID PRIMARY KEY,
    topic      TEXT NOT NULL,
    key        TEXT NOT NULL DEFAULT '',
    payload    JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS audit_events_topic_idx ON audit_events (topic, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_events;
-- +goose StatementEnd
