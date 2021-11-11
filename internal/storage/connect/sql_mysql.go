//go:build mysql
// +build mysql

package connect

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["mysql"] = mysqlConnect
}

func mysqlConnect(ctx context.Context, connURL string) (storage.Driver, error) {
	return sql.Open(ctx, "mysql", connURL)
}
