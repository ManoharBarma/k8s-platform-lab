-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS emi_installments (
    id        UUID PRIMARY KEY,
    loan_id   TEXT NOT NULL,
    number    INT NOT NULL,
    due_date  TIMESTAMPTZ NOT NULL,
    principal NUMERIC(38,2) NOT NULL,
    interest  NUMERIC(38,2) NOT NULL,
    total     NUMERIC(38,2) NOT NULL,
    balance   NUMERIC(38,2) NOT NULL,
    paid      BOOLEAN NOT NULL DEFAULT false,
    paid_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS emi_installments_loan_idx ON emi_installments (loan_id, number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS emi_installments;
-- +goose StatementEnd
