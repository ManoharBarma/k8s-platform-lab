-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS fx_rates (
    code    TEXT PRIMARY KEY,
    per_usd NUMERIC(38,6) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fx_rates;
-- +goose StatementEnd
