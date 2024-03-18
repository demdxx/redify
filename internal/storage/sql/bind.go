package sql

import (
	"context"
	"encoding/json"

	"github.com/demdxx/redify/internal/context/ctxlogger"
	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Bind struct {
	BindAbstract
	db               *sqlx.DB
	reorganizeNested bool
	minSizeOfRecord  int
}

// NewBind create new sql bind instance for the specified database
func NewBind(
	db *sqlx.DB,
	dbnum int,
	syntax Syntax,
	pattern, getQuery, listQuery, upsertQuery, delQuery string,
	datatypesMapping []storage.DatatypeMapper,
	reorganizeNested bool,
) *Bind {
	return &Bind{
		BindAbstract:     *NewBindAbstract(dbnum, syntax, pattern, getQuery, listQuery, upsertQuery, delQuery, datatypesMapping),
		db:               db,
		reorganizeNested: reorganizeNested,
		minSizeOfRecord:  10,
	}
}

// NewBindFromTableName create new sql bind instance for the specified database
func NewBindFromTableName(
	db *sqlx.DB,
	dbnum int,
	syntax Syntax,
	pattern, tableName, whereExt string,
	readonly bool,
	datatypesMapping []storage.DatatypeMapper,
	reorganizeNested bool,
) *Bind {
	return &Bind{
		BindAbstract:     *NewBindAbstractFromTableName(dbnum, syntax, pattern, tableName, whereExt, datatypesMapping, readonly),
		db:               db,
		reorganizeNested: reorganizeNested,
		minSizeOfRecord:  10,
	}
}

func (b *Bind) Get(ctx context.Context, ectx keypattern.ExecContext) (Record, error) {
	record := make(Record, b.minSizeOfRecord)
	rows, err := b.db.QueryxContext(ctx, b.GetQuery.String(), b.GetQuery.Args(ectx)...)
	ctxlogger.Get(ctx).Debug("Get",
		zap.Int("dbnum", b.DBNum),
		zap.String("query", b.GetQuery.String()),
		zap.Any("args", b.GetQuery.Args(ectx)),
		zap.Error(err),
	)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		err = rows.MapScan(record)
	}
	if len(record) != b.minSizeOfRecord {
		b.minSizeOfRecord = len(record)
	}
	if b.reorganizeNested {
		newRecord, err := record.ReorganizeNested()
		if err != nil {
			return nil, err
		}
		record = newRecord
	}

	if len(b.DatatypesMapping) > 0 {
		record, err = record.DatetypeCasting(b.DatatypesMapping...)
		if err != nil {
			return nil, err
		}
	}
	return record, err
}

func (b *Bind) List(ctx context.Context, ectx keypattern.ExecContext) ([]Record, error) {
	if b.ListQuery == nil {
		return nil, nil
	}
	res := make([]Record, 0, 10)
	rows, err := b.db.QueryxContext(ctx, b.ListQuery.String(), b.ListQuery.Args(ectx)...)
	ctxlogger.Get(ctx).Debug("List",
		zap.Int("dbnum", b.DBNum),
		zap.String("query", b.GetQuery.String()),
		zap.Any("args", b.GetQuery.Args(ectx)),
		zap.Error(err),
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		record := make(Record, b.minSizeOfRecord)
		if err = rows.MapScan(record); err != nil {
			return nil, err
		}
		if len(record) != b.minSizeOfRecord {
			b.minSizeOfRecord = len(record)
		}
		if b.reorganizeNested {
			newRecord, err := record.ReorganizeNested()
			if err != nil {
				return nil, err
			}
			record = newRecord
		}
		if len(b.DatatypesMapping) > 0 {
			record, err = record.DatetypeCasting(b.DatatypesMapping...)
			if err != nil {
				return nil, err
			}
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
	ctxlogger.Get(ctx).Debug("Upsert",
		zap.Int("dbnum", b.DBNum),
		zap.String("query", b.GetQuery.String()),
		zap.Any("args", b.GetQuery.Args(ectx)),
		zap.Error(err),
	)
	return err
}

func (b *Bind) Del(ctx context.Context, ectx keypattern.ExecContext) error {
	if b.DelQuery == nil {
		return storage.ErrReadOnly
	}
	_, err := b.db.ExecContext(ctx, b.DelQuery.String(), b.DelQuery.Args(ectx)...)
	ctxlogger.Get(ctx).Debug("Del",
		zap.Int("dbnum", b.DBNum),
		zap.String("query", b.GetQuery.String()),
		zap.Any("args", b.GetQuery.Args(ectx)),
		zap.Error(err),
	)
	return err
}
