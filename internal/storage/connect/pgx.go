//go:build pgx
// +build pgx

package connect

import (
	"github.com/demdxx/redify/internal/storage/pgx"
)

func init() {
	connectors["pgx"] = pgx.Open
	connectors["postgres"] = pgx.Open
}
