//go:build mssql || sqlserver
// +build mssql sqlserver

package connect

import (
	"context"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["mssql"] = mssqlConnect
	connectors["sqlserver"] = mssqlConnect
}

func mssqlConnect(ctx context.Context, connURL string) (storage.Driver, error) {
	return sql.Open(ctx, "sqlserver", connURL)
}
