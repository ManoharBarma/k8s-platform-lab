-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analytics_metrics (
    key   TEXT PRIMARY KEY,
    count BIGINT NOT NULL DEFAULT 0,
    sum   NUMERIC(38,2) NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analytics_metrics;
-- +goose StatementEnd
