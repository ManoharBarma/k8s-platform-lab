-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS search_index (
    id         UUID PRIMARY KEY,
    ref_id     TEXT NOT NULL,
    kind       TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS search_index_created_at_idx ON search_index (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS search_index;
-- +goose StatementEnd
