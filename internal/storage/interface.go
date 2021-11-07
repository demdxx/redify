package storage

import (
	"context"
	"errors"
	"io"
)

type Record map[string]interface{}

func (r Record) Get(key string) interface{} {
	if r == nil {
		return nil
	}
	return r[key]
}

var (
	ErrNotFound             = errors.New("not found")
	ErrNoKey                = errors.New("no key")
	ErrReadOnly             = errors.New("readonly access")
	ErrInvalidBindConfig    = errors.New("invalid bind config")
	ErrMethodIsNotSupported = errors.New("method is not supported")
)

type BindConfig struct {
	Pattern     string `json:"pattern" xml:"pattern" yaml:"pattern" toml:"pattern"`
	DBNum       int    `json:"dbnum" xml:"dbnum" yaml:"dbnum" toml:"dbnum"`
	TableName   string `json:"table_name" xml:"table_name" yaml:"table_name" toml:"table_name"`
	Readonly    bool   `json:"readonly" xml:"readonly" yaml:"readonly" toml:"readonly"`
	WhereExt    string `json:"where_ext" xml:"where_ext" yaml:"where_ext" toml:"where_ext"`
	GetQuery    string `json:"get_query" xml:"get_query" yaml:"get_query" toml:"get_query"`
	ListQuery   string `json:"list_query" xml:"list_query" yaml:"list_query" toml:"list_query"`
	UpsertQuery string `json:"upsert_query" xml:"upsert_query" yaml:"upsert_query" toml:"upsert_query"`
	DelQuery    string `json:"del_query" xml:"del_query" yaml:"del_query" toml:"del_query"`
}

// Driver storage description
type Driver interface {
	io.Closer
	Get(ctx context.Context, dbnum int, key string) ([]byte, error)
	Set(ctx context.Context, dbnum int, key string, value []byte) error
	Del(ctx context.Context, dbnum int, key string) error
	Keys(ctx context.Context, dbnum int, pattern string) ([]string, error)
	Bind(ctx context.Context, conf *BindConfig) error
}

type Cacher interface {
	io.Closer
	WithPrefix(prefix string) Cacher
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Del(ctx context.Context, key string) error
}
