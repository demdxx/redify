package proxy

import (
	"context"
	"errors"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/demdxx/redify/internal/context/ctxlogger"
	"github.com/demdxx/redify/internal/storage"
)

type notifyListener interface {
	ListenUpdateNotifies(ctx context.Context, chanelName string, notifyFnk func(ctx context.Context, key string)) error
}

type proxyStore struct {
	cache storage.Cacher
	store storage.Driver
}

// New proxy driver cache implementation
func New(ctx context.Context, cache storage.Cacher, store storage.Driver, notifyChannelName string) storage.Driver {
	if cache == nil {
		return store
	}
	prx := &proxyStore{
		cache: cache,
		store: store,
	}
	if notifier, _ := store.(notifyListener); notifier != nil && notifyChannelName != "" {
		go func() {
			ctxlogger.Get(ctx).Info("run notify listener")
			if err := notifier.ListenUpdateNotifies(ctx, notifyChannelName, prx.notifier); err != nil {
				ctxlogger.Get(ctx).Error("notification updates listener", zap.Error(err))
			}
		}()
	}
	return prx
}

func (d *proxyStore) Get(ctx context.Context, dbnum int, key string) ([]byte, error) {
	val, err := d.cache.Get(ctx, key)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		ctxlogger.Get(ctx).Debug("get value from cache",
			zap.String("key", key), zap.Int("dbnum", dbnum))
		return val, nil
	}
	if val, err = d.store.Get(ctx, dbnum, key); err != nil {
		return nil, err
	}
	ctxlogger.Get(ctx).Debug("get value from store",
		zap.String("key", key), zap.Int("dbnum", dbnum))
	if err = d.cache.Set(ctx, key, val); err != nil {
		return nil, err
	}
	return val, nil
}

func (d *proxyStore) Set(ctx context.Context, dbnum int, key string, value []byte) error {
	err := d.store.Set(ctx, dbnum, key, value)
	if err == nil {
		cerr := d.cache.Set(ctx, key, value)
		if cerr != nil {
			ctxlogger.Get(ctx).Error("cache set", zap.Error(err))
		}
	}
	return err
}

func (d *proxyStore) Del(ctx context.Context, dbnum int, key string) error {
	err := d.cache.Del(ctx, key)
	err = multierr.Append(err, d.store.Del(ctx, dbnum, key))
	return err
}

func (d *proxyStore) Keys(ctx context.Context, dbnum int, pattern string) ([]string, error) {
	return d.store.Keys(ctx, dbnum, pattern)
}

func (d *proxyStore) List(ctx context.Context, dbnum int, pattern string) ([]byte, error) {
	return d.store.List(ctx, dbnum, pattern)
}

func (d *proxyStore) Bind(ctx context.Context, conf *storage.BindConfig) error {
	return d.store.Bind(ctx, conf)
}

func (d *proxyStore) notifier(ctx context.Context, key string) {
	if err := d.cache.Del(ctx, key); err != nil {
		if err != storage.ErrNotFound && err != storage.ErrNoKey {
			ctxlogger.Get(ctx).Error("clear key cache", zap.String("key", key), zap.Error(err))
		}
	} else {
		ctxlogger.Get(ctx).Debug("clear key cache", zap.String("key", key))
	}
}

func (d *proxyStore) Close() error {
	err := d.cache.Close()
	err = multierr.Append(err, d.store.Close())
	return err
}
