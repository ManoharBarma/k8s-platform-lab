-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS kyc_verifications (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL,
    document_type TEXT NOT NULL,
    document_id   TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'PENDING',
    reason        TEXT NOT NULL DEFAULT '',
    submitted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS kyc_verifications_user_idx ON kyc_verifications (user_id);
CREATE INDEX IF NOT EXISTS kyc_verifications_status_idx ON kyc_verifications (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS kyc_verifications;
-- +goose StatementEnd
