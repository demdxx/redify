package simplecache

import (
	"context"
	"fmt"
	"testing"

	"github.com/demdxx/redify/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	var (
		ctx            = context.Background()
		cacheMain, err = New(10, 0)
		caches         = []storage.Cacher{cacheMain}
	)
	if !assert.NoError(t, err, "new cache object") {
		return
	}
	caches = append(caches, cacheMain.WithPrefix("cache1_"))
	caches = append(caches, cacheMain.WithPrefix("cache2_"))
	for i, cache := range caches {
		t.Run(fmt.Sprintf("cache_test_%d", i), func(t *testing.T) {
			_, err = cache.Get(ctx, "key1")
			assert.ErrorIs(t, err, storage.ErrNotFound)
			assert.NoError(t, cache.Set(ctx, "key1", []byte("val")), "set value")
			data, err := cache.Get(ctx, "key1")
			assert.NoError(t, err, "key must exist")
			assert.Equal(t, []byte("val"), data)
			assert.NoError(t, cache.Del(ctx, "key1"))
			assert.ErrorIs(t, cache.Del(ctx, "key_undefined"), storage.ErrNotFound)
			assert.NoError(t, cache.Close())
		})
	}
}
