package lrucache

import (
	"context"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/demdxx/redify/internal/cache"
	"github.com/demdxx/redify/internal/fasttime"
)

type item struct {
	value       []byte
	createdTime uint64
}

type lruCache struct {
	ttl    uint64 // in seconds
	prefix string
	cache  *lru.Cache[string, item]
}

// New LRU driver cache implementation
func New(size, ttl int) (*lruCache, error) {
	cache, err := lru.New[string, item](size)
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		ttl = 60
	}
	return &lruCache{
		ttl:   uint64(ttl),
		cache: cache,
	}, nil
}

func (d *lruCache) WithPrefix(prefix string) cache.Cacher {
	return &lruCache{
		ttl:    d.ttl,
		prefix: prefix,
		cache:  d.cache,
	}
}

func (d *lruCache) Get(ctx context.Context, key string) ([]byte, error) {
	key = d.prefix + key
	val, ok := d.cache.Get(key)
	if !ok {
		return nil, cache.ErrNotFound
	}
	if val.createdTime+d.ttl < fasttime.UnixTimestamp() {
		_ = d.cache.Remove(key)
		return nil, cache.ErrNotFound
	}
	return val.value, nil
}

func (d *lruCache) Set(ctx context.Context, key string, value []byte) error {
	_ = d.cache.Add(d.prefix+key, item{value: value, createdTime: fasttime.UnixTimestamp()})
	return nil
}

func (d *lruCache) Del(ctx context.Context, key string) error {
	if !d.cache.Remove(d.prefix + key) {
		return cache.ErrNotFound
	}
	return nil
}

func (d *lruCache) Close() error {
	d.cache.Purge()
	return nil
}
