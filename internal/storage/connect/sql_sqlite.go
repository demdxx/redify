//go:build sqlite || sqlite3
// +build sqlite sqlite3

package connect

import (
	"context"

	_ "github.com/mattn/go-sqlite3"

	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["sqlite"] = sqliteConnect
	connectors["sqlite3"] = sqliteConnect
}

func sqliteConnect(ctx context.Context, connURL string) (storage.Driver, error) {
	return sql.Open(ctx, "sqlite3", connURL)
}
