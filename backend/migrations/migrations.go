// Package migrations embeds the versioned SQL migration files so they can be
// applied automatically on boot without shipping the .sql files alongside the
// binary. Files follow the "<version>_<name>.sql" convention (e.g.
// "0001_init.sql") and are applied in ascending version order.
package migrations

import "embed"

// FS holds the embedded migration SQL files.
//
//go:embed *.sql
var FS embed.FS
