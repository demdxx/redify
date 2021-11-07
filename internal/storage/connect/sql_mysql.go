//go:build mysql

package connect

import (
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["mysql"] = sql.Open
}
