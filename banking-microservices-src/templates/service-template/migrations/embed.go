// Package migrations embeds this service's SQL migrations so they can be run at
// startup without shipping the .sql files separately.
package migrations

import "embed"

// FS holds the embedded goose migration files.
//
//go:embed *.sql
var FS embed.FS
