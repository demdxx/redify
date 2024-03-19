package simplecache

import (
	"context"
	"fmt"
	"testing"

	"github.com/demdxx/redify/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	var (
		ctx            = context.Background()
		cacheMain, err = New(10, 0)
		caches         = []cache.Cacher{cacheMain}
	)
	if !assert.NoError(t, err, "new cache object") {
		return
	}
	caches = append(caches,
		cacheMain.WithPrefix("cache1_"),
		cacheMain.WithPrefix("cache2_"))
	for i, cacheObj := range caches {
		t.Run(fmt.Sprintf("cache_test_%d", i), func(t *testing.T) {
			_, err = cacheObj.Get(ctx, "key1")
			assert.ErrorIs(t, err, cache.ErrNotFound)
			assert.NoError(t, cacheObj.Set(ctx, "key1", []byte("val")), "set value")
			data, err := cacheObj.Get(ctx, "key1")
			assert.NoError(t, err, "key must exist")
			assert.Equal(t, []byte("val"), data)
			assert.NoError(t, cacheObj.Del(ctx, "key1"))
			// assert.ErrorIs(t, cacheObj.Del(ctx, "key_undefined"), storage.ErrNotFound)
			assert.NoError(t, cacheObj.Close())
		})
	}
}
