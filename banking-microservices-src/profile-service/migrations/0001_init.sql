-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS profile_profiles (
    user_id       TEXT PRIMARY KEY,
    date_of_birth TEXT NOT NULL DEFAULT '',
    gender        TEXT NOT NULL DEFAULT '',
    address_line1 TEXT NOT NULL DEFAULT '',
    address_line2 TEXT NOT NULL DEFAULT '',
    city          TEXT NOT NULL DEFAULT '',
    state         TEXT NOT NULL DEFAULT '',
    country       TEXT NOT NULL DEFAULT '',
    postal_code   TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile_profiles;
-- +goose StatementEnd
