//go:build clickhouse
// +build clickhouse

package connect

import (
	"context"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

func init() {
	connectors["clickhouse"] = func(ctx context.Context, connURL string) (storage.Driver, error) {
		return sql.Open(ctx, "clickhouse", connURL)
	}
}
