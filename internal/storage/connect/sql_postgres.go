//go:build (postgres || pgsql) && !pgx
// +build postgres pgsql
// +build !pgx

package connect

import (
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["postgres"] = sql.Open
}
