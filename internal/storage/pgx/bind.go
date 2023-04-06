package pgx

import (
	"context"
	"encoding/json"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

type (
	Record     = storage.Record
	DataFields = sql.DataFields
	WhereStmt  = sql.WhereStmt
	Syntax     = sql.Syntax
)

type pgpoolIface interface {
	pgxscan.Querier
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Bind struct {
	conn pgpoolIface
	sql.BindAbstract
	minSizeOfRecord int
}

// NewBind create new sql bind instance for the specified database
func NewBind(
	conn pgpoolIface,
	dbnum int,
	syntax Syntax,
	pattern, getQuery, listQuery, upsertQuery, delQuery string,
	datatypesMapping []storage.DatatypeMapper,
) *Bind {
	return &Bind{
		BindAbstract:    *sql.NewBindAbstract(dbnum, syntax, pattern, getQuery, listQuery, upsertQuery, delQuery, datatypesMapping),
		conn:            conn,
		minSizeOfRecord: 10,
	}
}

// NewBindFromTableName create new sql bind instance for the specified database
func NewBindFromTableName(
	conn pgpoolIface,
	dbnum int,
	syntax Syntax,
	pattern, tableName, whereExt string,
	datatypesMapping []storage.DatatypeMapper,
	readonly bool,
) *Bind {
	return &Bind{
		BindAbstract:    *sql.NewBindAbstractFromTableName(dbnum, syntax, pattern, tableName, whereExt, datatypesMapping, readonly),
		conn:            conn,
		minSizeOfRecord: 10,
	}
}

func (b *Bind) Get(ctx context.Context, ectx keypattern.ExecContext) (Record, error) {
	rows, err := b.conn.Query(ctx, b.GetQuery.String(), b.GetQuery.Args(ectx)...)
	if err != nil {
		return nil, err
	}
	record := make(Record, b.minSizeOfRecord)
	err = pgxscan.ScanOne(&record, rows)
	if err != nil {
		return nil, err
	}
	if len(b.DatatypesMapping) > 0 {
		record, err = record.DatetypeCasting(b.DatatypesMapping...)
		if err != nil {
			return nil, err
		}
	}
	if len(record) != b.minSizeOfRecord {
		b.minSizeOfRecord = len(record)
	}
	return record, err
}

func (b *Bind) List(ctx context.Context, ectx keypattern.ExecContext) ([]Record, error) {
	res := make([]Record, 0, 10)
	err := pgxscan.Select(ctx, b.conn, &res, b.ListQuery.String(), b.ListQuery.Args(ectx)...)
	if err != nil {
		return nil, err
	}
	if len(b.DatatypesMapping) > 0 {
		for i, record := range res {
			record, err = record.DatetypeCasting(b.DatatypesMapping...)
			if err != nil {
				return nil, err
			}
			res[i] = record
		}
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
	_, err := b.conn.Exec(ctx, b.UpsertQuery.String(), b.UpsertQuery.Args(ectx)...)
	return err
}

func (b *Bind) Del(ctx context.Context, ectx keypattern.ExecContext) error {
	if b.DelQuery == nil {
		return storage.ErrReadOnly
	}
	_, err := b.conn.Exec(ctx, b.DelQuery.String(), b.DelQuery.Args(ectx)...)
	return err
}
