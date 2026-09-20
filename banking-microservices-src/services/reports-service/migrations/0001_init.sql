-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reports_daily (
    date         TEXT NOT NULL,
    currency     TEXT NOT NULL,
    count        BIGINT NOT NULL DEFAULT 0,
    total_amount NUMERIC(38,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (date, currency)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reports_daily;
-- +goose StatementEnd
