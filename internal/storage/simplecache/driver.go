package simplecache

import (
	"context"
	"time"

	"github.com/ReneKroon/ttlcache/v2"

	"github.com/demdxx/redify/internal/storage"
)

type simpleCache struct {
	prefix string
	cache  *ttlcache.Cache
}

// New simple driver cache implementation
func New(size, ttl int) (storage.Cacher, error) {
	cache := ttlcache.NewCache()
	if ttl <= 0 {
		ttl = 60
	}
	err := cache.SetTTL(time.Duration(ttl) * time.Second)
	if err != nil {
		return nil, err
	}
	cache.SetCacheSizeLimit(size)
	return &simpleCache{
		cache: cache,
	}, nil
}

func (d *simpleCache) WithPrefix(prefix string) storage.Cacher {
	return &simpleCache{
		prefix: prefix,
		cache:  d.cache,
	}
}

func (d *simpleCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := d.cache.Get(d.prefix + key)
	if err != nil {
		if err == ttlcache.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return val.([]byte), nil
}

func (d *simpleCache) Set(ctx context.Context, key string, value []byte) error {
	return d.cache.Set(d.prefix+key, value)
}

func (d *simpleCache) Del(ctx context.Context, key string) error {
	if err := d.cache.Remove(d.prefix + key); err != nil {
		if err == ttlcache.ErrNotFound {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func (d *simpleCache) Close() error {
	return d.cache.Purge()
}
