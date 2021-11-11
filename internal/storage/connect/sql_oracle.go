//go:build oracle
// +build oracle

package connect

import (
	"context"

	_ "github.com/godror/godror"

	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["oracle"] = godrorConnect
}

func godrorConnect(ctx context.Context, connURL string) (storage.Driver, error) {
	return sql.Open(ctx, "godror", connURL)
}
