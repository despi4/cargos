package migrations

import "embed"

// FS содержит все sql файлы миграций, встроенные в бинарник
//go:embed *.sql
var FS embed.FS
