package pubstream

import (
	"context"
	"testing"

	"github.com/demdxx/redify/internal/storage"
	"github.com/geniusrabbit/notificationcenter/dummy"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	var (
		ctx       = context.Background()
		dummyPub  = dummy.Publisher{}
		dr        = &driver{publisher: dummyPub}
		bind, err = dr.bindByKey("no_key", 0)
	)
	assert.ErrorIs(t, err, storage.ErrNoKey)
	assert.Nil(t, bind, "undefined must be nil")
	assert.False(t, dr.SupportCache())
	assert.NoError(t, dr.Bind(ctx, &storage.BindConfig{Pattern: "some_key", DBNum: 1}))
	assert.NoError(t, dr.Bind(ctx, &storage.BindConfig{Pattern: "key1", DBNum: 1}))
	assert.NoError(t, dr.Bind(ctx, &storage.BindConfig{Pattern: "key2", DBNum: 1}))
	assert.NoError(t, dr.Bind(ctx, &storage.BindConfig{Pattern: "key3", DBNum: 3}))
	assert.Equal(t, 4, len(dr.binds))

	bind, err = dr.bindByKey("key1", 0)
	assert.ErrorIs(t, err, storage.ErrNoKey)
	assert.Nil(t, bind, "undefined must be nil")

	bind, err = dr.bindByKey("key1", 1)
	assert.NoError(t, err)
	if assert.NotNil(t, bind) {
		assert.Equal(t, "key1", bind.key)
		assert.Equal(t, 1, bind.dbnum)
	}

	assert.ErrorIs(t, dr.Set(ctx, 1, "undefined_key", []byte("")), storage.ErrNoKey)
	assert.NoError(t, dr.Set(ctx, 1, "key1", []byte("")))
	data, err := dr.Get(ctx, 1, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("1"), data)

	keys, err := dr.Keys(ctx, 1, "key*")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2"}, keys)

	assert.ErrorIs(t, dr.Del(ctx, 1, "key1"), storage.ErrMethodIsNotSupported)
	assert.NoError(t, dr.Close())
}
