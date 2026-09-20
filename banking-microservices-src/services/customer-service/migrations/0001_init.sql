-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_customers (
    id         UUID PRIMARY KEY,
    user_id    TEXT NOT NULL,
    email      TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    last_name  TEXT NOT NULL DEFAULT '',
    phone      TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT customer_customers_user_id_key UNIQUE (user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_customers;
-- +goose StatementEnd
