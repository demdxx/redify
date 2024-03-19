package cache

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoKey    = errors.New("no key")
)

// Cacher manage interface
type Cacher interface {
	io.Closer
	WithPrefix(prefix string) Cacher
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Del(ctx context.Context, key string) error
}
