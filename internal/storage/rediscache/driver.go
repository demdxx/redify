package rediscache

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/redify/internal/storage"
)

type simpleCache struct {
	conn   *retainConnect
	ttl    time.Duration
	prefix string
}

// New redis driver cache implementation
func New(connect string, ttl time.Duration) (storage.Cacher, error) {
	urlObj, err := url.Parse(connect)
	if err != nil {
		return nil, err
	}
	password, _ := urlObj.User.Password()
	username := urlObj.User.Username()
	rdb := redis.NewClient(&redis.Options{
		Network:            _schema(urlObj.Scheme),
		Addr:               urlObj.Host,
		Username:           username,
		Password:           password,
		DB:                 gocast.Int(strings.TrimLeft(urlObj.Path, "/")),
		MaxRetries:         gocast.Int(urlObj.Query().Get("max_retries")),
		MinRetryBackoff:    _duration(urlObj.Query().Get("min_retry_backoff"), 0),
		MaxRetryBackoff:    _duration(urlObj.Query().Get("max_retry_backoff"), 0),
		DialTimeout:        _duration(urlObj.Query().Get("dial_timeout"), 0),
		ReadTimeout:        _duration(urlObj.Query().Get("read_timeout"), 0),
		WriteTimeout:       _duration(urlObj.Query().Get("write_timeout"), 0),
		PoolFIFO:           gocast.Bool(urlObj.Query().Get("pool_fifo")),
		PoolSize:           gocast.Int(urlObj.Query().Get("pool_size")),
		MinIdleConns:       gocast.Int(urlObj.Query().Get("min_idle_conns")),
		MaxConnAge:         _duration(urlObj.Query().Get("max_conn_age"), 0),
		PoolTimeout:        _duration(urlObj.Query().Get("pool_timeout"), 0),
		IdleTimeout:        _duration(urlObj.Query().Get("idle"), 0),
		IdleCheckFrequency: _duration(urlObj.Query().Get("idle_check_frequency"), 0),
	})
	return newFromConnect(
		&retainConnect{Cmdable: rdb},
		_or(ttl, _duration(urlObj.Query().Get("ttl"), time.Second*60)),
		"",
	), nil
}

func newFromConnect(conn *retainConnect, ttl time.Duration, prefix string) storage.Cacher {
	return &simpleCache{
		conn:   conn,
		ttl:    ttl,
		prefix: prefix,
	}
}

func (d *simpleCache) WithPrefix(prefix string) storage.Cacher {
	return newFromConnect(d.conn.inc(), d.ttl, prefix)
}

func (d *simpleCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := d.conn.GetEx(ctx, d.prefix+key, d.ttl).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, storage.ErrNotFound
		}
		return nil, _err(err)
	}
	return []byte(val), nil
}

func (d *simpleCache) Set(ctx context.Context, key string, value []byte) error {
	return _err(d.conn.SetEX(ctx, d.prefix+key, value, d.ttl).Err())
}

func (d *simpleCache) Del(ctx context.Context, key string) error {
	return _err(d.conn.Del(ctx, d.prefix+key).Err())
}

func (d *simpleCache) Close() error {
	return d.conn.Close()
}

func _err(err error) error {
	if err == redis.Nil {
		return storage.ErrNotFound
	}
	return err
}

func _duration(tm string, def time.Duration) time.Duration {
	v, _ := time.ParseDuration(tm)
	if v == 0 {
		return def
	}
	return v
}

func _or[T gocast.Numeric](a, b T) T {
	if a <= 0 {
		return b
	}
	return a
}

func _schema(schema string) string {
	switch sch := strings.ToLower(schema); sch {
	case "redis", "tcp":
		return "tcp"
	case "redis-unix", "unix":
		return "unix"
	default:
		return sch
	}
}
