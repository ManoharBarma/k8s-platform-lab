-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS document_documents (
    id           UUID PRIMARY KEY,
    owner_id     TEXT NOT NULL,
    type         TEXT NOT NULL,
    key          TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS document_documents_owner_idx ON document_documents (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS document_documents;
-- +goose StatementEnd
