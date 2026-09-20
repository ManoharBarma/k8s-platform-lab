-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS authz_roles (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS authz_permissions (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS authz_role_permissions (
    role       TEXT NOT NULL,
    permission TEXT NOT NULL,
    PRIMARY KEY (role, permission)
);
CREATE TABLE IF NOT EXISTS authz_user_roles (
    subject TEXT NOT NULL,
    role    TEXT NOT NULL,
    PRIMARY KEY (subject, role)
);
CREATE INDEX IF NOT EXISTS authz_user_roles_subject_idx ON authz_user_roles (subject);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS authz_user_roles;
DROP TABLE IF EXISTS authz_role_permissions;
DROP TABLE IF EXISTS authz_permissions;
DROP TABLE IF EXISTS authz_roles;
-- +goose StatementEnd
