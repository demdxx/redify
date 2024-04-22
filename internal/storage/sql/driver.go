package sql

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
)

type sqlStore struct {
	driverName string
	db         *sqlx.DB
	binds      []*Bind
	syntax     Syntax
}

// Open sql driver connect
func Open(ctx context.Context, driver, connURL string) (storage.Driver, error) {
	db, err := sqlx.Open(driver, connURL)
	if err != nil {
		return nil, err
	}
	var syntax Syntax
	switch driver {
	case "postgres", "postgresql", "pgx":
		syntax = NewAbstractSyntax(`"`)
	case "mysql":
		syntax = NewMysqlSyntax()
	case "sqlite", "sqlite3":
		syntax = NewAbstractSyntax(`"`)
	case "mssql", "sqlserver":
		// TODO: Upsert query
		syntax = NewAbstractSyntax(`"`)
	case "clickhouse":
		syntax = NewAbstractSyntax("`")
	default:
		syntax = NewAbstractSyntax(`"`)
	}
	return &sqlStore{db: db, driverName: driver, syntax: syntax}, nil
}

func (dr *sqlStore) Get(ctx context.Context, dbnum int, key string) ([]byte, error) {
	ectx := keypattern.ExecContext{}
	bind, err := dr.bindByKey(key, dbnum, ectx)
	if err != nil {
		return nil, err
	}
	rec, err := bind.Get(ctx, ectx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rec)
}

func (dr *sqlStore) Set(ctx context.Context, dbnum int, key string, value []byte) error {
	ectx := keypattern.ExecContext{}
	bind, err := dr.bindByKey(key, dbnum, ectx)
	if err != nil {
		return err
	}
	return bind.Upsert(ctx, ectx, value)
}

func (dr *sqlStore) Del(ctx context.Context, dbnum int, key string) error {
	ectx := keypattern.ExecContext{}
	bind, err := dr.bindByKey(key, dbnum, ectx)
	if err != nil {
		return err
	}
	return bind.Del(ctx, ectx)
}

func (dr *sqlStore) Keys(ctx context.Context, dbnum int, pattern string) ([]string, error) {
	var (
		keys   []string
		hasKey bool
	)
	for _, bind := range dr.binds {
		ectx := keypattern.ExecContext{}
		if bind.DBNum != dbnum || !bind.MatchPattern(pattern, ectx) {
			continue
		}
		hasKey = true
		res, err := bind.List(ctx, ectx)
		if err != nil {
			return nil, err
		}
		if keys == nil {
			keys = make([]string, 0, len(res))
		}
		for _, r := range res {
			keys = append(keys, bind.Pattern.Format(r))
		}
	}
	if !hasKey {
		return nil, storage.ErrNoKey
	}
	return keys, nil
}

func (dr *sqlStore) Bind(ctx context.Context, conf *storage.BindConfig) error {
	var bind *Bind
	if conf.GetQuery != "" {
		bind = NewBind(dr.db, conf.DBNum, dr.syntax,
			conf.Pattern, conf.GetQuery, conf.ListQuery, conf.UpsertQuery, conf.DelQuery, conf.DatatypeMapping, conf.ReorganizeNested)
	} else if conf.TableName != "" {
		bind = NewBindFromTableName(dr.db, conf.DBNum, dr.syntax,
			conf.Pattern, conf.TableName, conf.WhereExt, conf.Readonly, conf.DatatypeMapping, conf.ReorganizeNested)
	} else {
		return storage.ErrInvalidBindConfig
	}
	bind.driverName = dr.driverName
	dr.binds = append(dr.binds, bind)
	return nil
}

func (dr *sqlStore) Close() error {
	return dr.db.Close()
}

func (dr *sqlStore) bindByKey(key string, dbnum int, ectx keypattern.ExecContext) (*Bind, error) {
	for _, b := range dr.binds {
		if b.DBNum == dbnum && b.MatchKey(key, ectx) {
			return b, nil
		}
	}
	return nil, storage.ErrNoKey
}
