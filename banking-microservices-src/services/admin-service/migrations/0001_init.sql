-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admin_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admin_settings;
-- +goose StatementEnd
