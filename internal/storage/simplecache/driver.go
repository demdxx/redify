package simplecache

import (
	"context"
	"errors"
	"time"

	"github.com/jellydator/ttlcache/v3"

	"github.com/demdxx/redify/internal/storage"
)

var errSaveItem = errors.New(`undefined error of item savings`)

type simpleCache struct {
	prefix string
	ttl    int
	size   int
	cache  *ttlcache.Cache[string, []byte]
}

// NewCache simple driver cache implementation
func NewCache(size, ttl int, prefix string) (storage.Cacher, error) {
	if ttl <= 0 {
		ttl = 60
	}
	cache := ttlcache.New(
		ttlcache.WithTTL[string, []byte](time.Duration(ttl)*time.Second),
		ttlcache.WithCapacity[string, []byte](uint64(size)),
	)
	// Run automatic cleanup
	go cache.Start()
	return &simpleCache{
		prefix: prefix,
		cache:  cache,
		ttl:    ttl,
		size:   size,
	}, nil
}

// New simple driver cache implementation
func New(size, ttl int) (storage.Cacher, error) {
	return NewCache(size, ttl, "")
}

func (d *simpleCache) WithPrefix(prefix string) storage.Cacher {
	cache, err := NewCache(d.size, d.ttl, prefix)
	if err != nil {
		panic(err)
	}
	return cache
}

func (d *simpleCache) Get(ctx context.Context, key string) ([]byte, error) {
	val := d.cache.Get(d.prefix + key)
	if val == nil {
		return nil, storage.ErrNotFound
	}
	return val.Value(), nil
}

func (d *simpleCache) Set(ctx context.Context, key string, value []byte) error {
	if it := d.cache.Set(d.prefix+key, value, ttlcache.DefaultTTL); it == nil {
		return errSaveItem
	}
	return nil
}

func (d *simpleCache) Del(ctx context.Context, key string) error {
	d.cache.Delete(d.prefix + key)
	return nil
}

func (d *simpleCache) Close() error {
	d.cache.Stop()
	d.cache.DeleteAll()
	return nil
}
