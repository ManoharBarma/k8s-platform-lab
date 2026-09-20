package postgres

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // register the "pgx" database/sql driver
	"github.com/pressly/goose/v3"
)

// Migrate runs goose "up" migrations from an embedded filesystem. Each service
// embeds its own migrations/ directory and calls this at startup.
//
// service names the calling service and is used to derive a PER-SERVICE goose
// version table. This is essential because the whole platform shares ONE
// database: with a single shared version table, goose would treat every
// service's "0001" migration as the same version and skip all but the first.
// A dedicated table per service keeps their migration histories independent.
func Migrate(dsn string, migrations fs.FS, dir, service string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	versionTable := "goose_db_version"
	if service != "" {
		versionTable = sanitize(service) + "_goose_version"
	}
	goose.SetTableName(versionTable)
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// sanitize makes a service name safe as a SQL identifier.
func sanitize(s string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(s)
}
