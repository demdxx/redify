package pubstream

import (
	"context"
	"io"
	"path/filepath"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/redify/internal/storage"
	nc "github.com/geniusrabbit/notificationcenter/v2"
)

// Driver of the abstract stream publisher
type driver struct {
	publisher nc.Publisher
	binds     []*bind
}

func Open(ctx context.Context, connURL string) (storage.Driver, error) {
	pub, err := connect(ctx, connURL)
	if err != nil {
		return nil, err
	}
	return &driver{publisher: pub}, nil
}

func (dr *driver) Get(ctx context.Context, dbnum int, key string) ([]byte, error) {
	bind, err := dr.bindByKey(key, dbnum)
	if err != nil {
		return nil, err
	}
	return []byte(gocast.Str(bind.CountPubs())), nil
}

func (dr *driver) Set(ctx context.Context, dbnum int, key string, value []byte) error {
	bind, err := dr.bindByKey(key, dbnum)
	if err != nil {
		return err
	}
	return bind.Publish(ctx, dr.publisher, value)
}

func (dr *driver) Del(ctx context.Context, dbnum int, key string) error {
	_, err := dr.bindByKey(key, dbnum)
	if err == nil {
		return storage.ErrMethodIsNotSupported
	}
	return nil
}

func (dr *driver) Keys(ctx context.Context, dbnum int, pattern string) ([]string, error) {
	keys := make([]string, 0, len(dr.binds))
	for _, bind := range dr.binds {
		if bind.dbnum != dbnum {
			continue
		}
		if matched, _ := filepath.Match(pattern, bind.key); !matched {
			continue
		}
		keys = append(keys, bind.key)
	}
	return keys, nil
}

func (dr *driver) List(ctx context.Context, dbnum int, pattern string) ([]storage.Record, error) {
	return nil, nil // storage.ErrMethodIsNotSupported
}

func (dr *driver) Bind(ctx context.Context, conf *storage.BindConfig) error {
	dr.binds = append(dr.binds, &bind{
		dbnum: conf.DBNum,
		key:   conf.Pattern,
	})
	return nil
}

func (dr *driver) Close() error {
	if cl, _ := dr.publisher.(io.Closer); cl != nil {
		return cl.Close()
	}
	return nil
}

func (dr *driver) SupportCache() bool {
	return false
}

func (dr *driver) bindByKey(key string, dbnum int) (*bind, error) {
	for _, b := range dr.binds {
		if b.dbnum == dbnum && b.key == key {
			return b, nil
		}
	}
	return nil, storage.ErrNoKey
}
