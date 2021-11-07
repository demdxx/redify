//go:build sqlite || sqlite3
// +build sqlite sqlite3

package connect

import (
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["sqlite"] = sql.Open
	connectors["sqlite3"] = sql.Open
}
