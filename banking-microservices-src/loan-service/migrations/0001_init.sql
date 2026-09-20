-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS loan_loans (
    id            UUID PRIMARY KEY,
    borrower_id   TEXT NOT NULL,
    account_id    UUID NOT NULL,
    principal     NUMERIC(38,2) NOT NULL,
    currency      TEXT NOT NULL DEFAULT 'USD',
    annual_rate   DOUBLE PRECISION NOT NULL DEFAULT 0,
    tenure_months INT NOT NULL,
    emi_amount    NUMERIC(38,2) NOT NULL DEFAULT 0,
    outstanding   NUMERIC(38,2) NOT NULL DEFAULT 0,
    status        TEXT NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    disbursed_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS loan_borrower_idx ON loan_loans (borrower_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS loan_loans;
-- +goose StatementEnd
