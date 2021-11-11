package sql

import (
	"context"
	"encoding/json"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/jmoiron/sqlx"
)

type Bind struct {
	BindAbstract
	db *sqlx.DB
}

func NewBind(db *sqlx.DB, dbnum int, syntax Syntax, pattern, getQuery, listQuery, upsertQuery, delQuery string) *Bind {
	return &Bind{
		BindAbstract: *NewBindAbstract(dbnum, syntax, pattern, getQuery, listQuery, upsertQuery, delQuery),
		db:           db,
	}
}

func NewBindFromTableName(db *sqlx.DB, dbnum int, syntax Syntax, pattern, tableName, whereExt string, readonly bool) *Bind {
	return &Bind{
		BindAbstract: *NewBindAbstractFromTableName(dbnum, syntax, pattern, tableName, whereExt, readonly),
		db:           db,
	}
}

func (b *Bind) Get(ctx context.Context, ectx keypattern.ExecContext) (Record, error) {
	record := make(Record, 10)
	rows, err := b.db.QueryxContext(ctx, b.GetQuery.String(), b.GetQuery.Args(ectx)...)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		err = rows.MapScan(record)
	}
	return record, err
}

func (b *Bind) List(ctx context.Context, ectx keypattern.ExecContext) ([]Record, error) {
	res := make([]Record, 0, 10)
	rows, err := b.db.QueryxContext(ctx, b.ListQuery.String(), b.ListQuery.Args(ectx)...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		record := make(Record, 10)
		if err = rows.MapScan(record); err != nil {
			return nil, err
		}
		res = append(res, record)
	}
	return res, err
}

func (b *Bind) Upsert(ctx context.Context, ectx keypattern.ExecContext, value []byte) error {
	if b.UpsertQuery == nil {
		return storage.ErrReadOnly
	}
	if ectx == nil {
		ectx = make(keypattern.ExecContext, 10)
	}
	var values keypattern.ExecContext
	if err := json.Unmarshal(value, &values); err != nil {
		return err
	}
	for k, v := range values {
		ectx[k] = v
	}
	_, err := b.db.ExecContext(ctx, b.UpsertQuery.String(), b.UpsertQuery.Args(ectx)...)
	return err
}

func (b *Bind) Del(ctx context.Context, ectx keypattern.ExecContext) error {
	if b.DelQuery == nil {
		return storage.ErrReadOnly
	}
	_, err := b.db.ExecContext(ctx, b.DelQuery.String(), b.DelQuery.Args(ectx)...)
	return err
}
