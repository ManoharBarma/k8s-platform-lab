-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth_users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    full_name     TEXT NOT NULL DEFAULT '',
    roles         TEXT NOT NULL DEFAULT 'customer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT auth_users_email_key UNIQUE (email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_users;
-- +goose StatementEnd
