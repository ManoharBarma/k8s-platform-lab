-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tmpl_resources (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tmpl_resources;
-- +goose StatementEnd
