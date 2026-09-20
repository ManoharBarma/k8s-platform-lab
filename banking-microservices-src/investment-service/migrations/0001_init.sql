-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS inv_holdings (
    id         UUID NOT NULL,
    owner_id   TEXT NOT NULL,
    symbol     TEXT NOT NULL,
    units      NUMERIC(38,4) NOT NULL DEFAULT 0,
    avg_price  NUMERIC(38,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_id, symbol)
);
CREATE INDEX IF NOT EXISTS inv_owner_idx ON inv_holdings (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inv_holdings;
-- +goose StatementEnd
