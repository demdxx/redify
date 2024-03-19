package rediscache

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/demdxx/redify/internal/cache"
	"github.com/demdxx/redify/internal/storage"
	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

// newTestRedis returns a redis.Cmdable.
func newTestRedis(mr *miniredis.Miniredis) *redismock.ClientMock {
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return redismock.NewNiceMock(client)
}

func TestDriver(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	var (
		mockCli   = newTestRedis(mr)
		ctx       = context.Background()
		cacheMain = newFromConnect(&retainConnect{Cmdable: mockCli}, 0, "")
		caches    = []cache.Cacher{cacheMain}
	)
	if !assert.NoError(t, err, "new cache object") {
		return
	}
	caches = append(caches,
		cacheMain.WithPrefix("cache1_"),
		cacheMain.WithPrefix("cache2_"))
	for i, cache := range caches {
		t.Run(fmt.Sprintf("cache_test_%d", i), func(t *testing.T) {
			_, err = cache.Get(ctx, "key1")
			assert.ErrorIs(t, err, storage.ErrNotFound)
			assert.NoError(t, cache.Set(ctx, "key1", []byte("val")), "set value")
			data, err := cache.Get(ctx, "key1")
			assert.NoError(t, err, "key must exist")
			assert.Equal(t, []byte("val"), data)
			assert.NoError(t, cache.Del(ctx, "key1"))
			// assert.ErrorIs(t, cache.Del(ctx, "key_undefined"), storage.ErrNotFound)
			assert.NoError(t, cache.Close())
		})
	}
}
