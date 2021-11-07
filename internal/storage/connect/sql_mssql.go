//go:build mssql

package connect

import (
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["mssql"] = sql.Open
}
