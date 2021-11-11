//go:build (postgres || pgsql) && !pgx
// +build postgres pgsql
// +build !pgx

package connect

import (
	"context"

	_ "github.com/lib/pq"

	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["postgres"] = func(ctx context.Context, connURL string) (storage.Driver, error) {
		return sql.Open(ctx, "postgres", connURL)
	}
}
