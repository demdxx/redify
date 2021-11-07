package multistore

import (
	"context"

	"github.com/demdxx/redify/internal/storage"
	"go.uber.org/multierr"
)

type Driver struct {
	stores []storage.Driver
}

func New(stores ...storage.Driver) *Driver {
	return &Driver{stores: stores}
}

func (d *Driver) Get(ctx context.Context, dbnum int, key string) (value []byte, err error) {
	for _, st := range d.stores {
		value, err := st.Get(ctx, dbnum, key)
		if err == storage.ErrNoKey {
			continue
		}
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	return nil, storage.ErrNoKey
}

func (d *Driver) Set(ctx context.Context, dbnum int, key string, value []byte) (err error) {
	for _, st := range d.stores {
		serr := st.Set(ctx, dbnum, key, value)
		if serr == storage.ErrNoKey {
			continue
		}
		if serr != nil {
			err = multierr.Append(err, serr)
		}
	}
	return err
}

func (d *Driver) Del(ctx context.Context, dbnum int, key string) (err error) {
	for _, st := range d.stores {
		serr := st.Del(ctx, dbnum, key)
		if serr == storage.ErrNoKey {
			continue
		}
		if serr != nil {
			err = multierr.Append(err, serr)
		}
	}
	return err
}

func (d *Driver) Keys(ctx context.Context, dbnum int, pattern string) (keys []string, err error) {
	for _, st := range d.stores {
		skeys, serr := st.Keys(ctx, dbnum, pattern)
		if serr == storage.ErrNoKey {
			continue
		}
		if serr != nil {
			err = multierr.Append(err, serr)
		}
		keys = append(keys, skeys...)
	}
	return keys, err
}

func (d *Driver) Bind(ctx context.Context, conf *storage.BindConfig) error {
	return storage.ErrMethodIsNotSupported
}

func (d *Driver) Close() (err error) {
	for _, st := range d.stores {
		err = multierr.Append(err, st.Close())
	}
	return err
}
