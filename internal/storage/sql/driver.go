package sql

import (
	"context"

	"github.com/demdxx/redify/internal/storage"
)

type sqlStore struct {
}

// Open sql driver connect
func Open(ctx context.Context, connURL string) (storage.Driver, error) {
	return &sqlStore{}, nil
}

func (d *sqlStore) Get(ctx context.Context, dbnum int, key string) ([]byte, error) {
	return nil, nil
}

func (d *sqlStore) Set(ctx context.Context, dbnum int, key string, value []byte) error {
	return nil
}

func (d *sqlStore) Del(ctx context.Context, dbnum int, key string) error {
	return nil
}

func (d *sqlStore) Keys(ctx context.Context, dbnum int, pattern string) ([]string, error) {
	return nil, nil
}

func (d *sqlStore) Bind(ctx context.Context, conf *storage.BindConfig) error {
	return nil
}

func (d *sqlStore) Close() error {
	return nil
}
