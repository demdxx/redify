package pgx

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

type (
	Record = storage.Record
	query  = sql.Query
)

type pgpoolIface interface {
	pgxscan.Querier
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Bind struct {
	conn        pgpoolIface
	dbnum       int
	pattern     *keypattern.Pattern
	getQuery    *query
	listQuery   *query
	upsertQuery *query
	delQuery    *query
}

func NewBind(conn pgpoolIface, dbnum int, pattern, getQuery, listQuery, upsertQuery, delQuery string) *Bind {
	return &Bind{
		conn:        conn,
		dbnum:       dbnum,
		pattern:     keypattern.NewPatternFromExpression(pattern),
		getQuery:    sql.ParseQuery(getQuery),
		listQuery:   sql.ParseQuery(listQuery),
		upsertQuery: sql.ParseQuery(upsertQuery),
		delQuery:    sql.ParseQuery(delQuery),
	}
}

func NewBindFromTableName(conn pgpoolIface, dbnum int, pattern, tableName, whereExt string, readonly bool) *Bind {
	var (
		ptrObj             = keypattern.NewPatternFromExpression(pattern)
		keyFields          = ptrObj.Keys()
		whereConds         = make([]string, 0, len(keyFields))
		insertValues       = make([]string, 0, len(keyFields))
		setValues          = make([]string, 0, len(keyFields))
		fields             = strings.Join(keyFields, ", ")
		whereStatement     = ""
		whereListStatement = ""
		delQyeryObj        *query
		insertQyeryObj     *query
	)
	if keys := ptrObj.Keys(); len(keys) > 0 {
		whereStatement = " WHERE "
		for _, key := range keys {
			whereConds = append(whereConds, key+"={{"+key+"}}")
			insertValues = append(insertValues, "{{"+key+"}}")
		}
		setValues = whereConds
		whereStatement += strings.Join(whereConds, " AND ")
	}
	if whereExt != "" {
		if whereStatement == "" {
			whereStatement = " WHERE "
		} else {
			whereStatement += " AND "
		}
		whereListStatement = " WHERE " + whereExt
		whereStatement += whereExt
	}
	if !readonly {
		delQyeryObj = sql.ParseQuery("DELETE FROM " + tableName + whereStatement)
		insertQyeryObj = sql.ParseQuery("INSERT INTO " + tableName + "(" + fields + ")" +
			" VALUES(" + strings.Join(insertValues, ", ") + ")" +
			" ON CONFLICT (" + fields + ") DO UPDATE SET " + strings.Join(setValues, ","))
	}
	return &Bind{
		conn:        conn,
		dbnum:       dbnum,
		pattern:     ptrObj,
		getQuery:    sql.ParseQuery("SELECT * FROM " + tableName + whereStatement + " LIMIT 1"),
		listQuery:   sql.ParseQuery("SELECT * FROM " + tableName + whereListStatement),
		delQuery:    delQyeryObj,
		upsertQuery: insertQyeryObj,
	}
}

func (b *Bind) matchKey(key string, ectx keypattern.ExecContext) bool {
	return b.pattern.Match(key, ectx)
}

func (b *Bind) matchPattern(pt string, ectx keypattern.ExecContext) bool {
	ok, _ := filepath.Match(pt, b.pattern.String())
	return ok
}

func (b *Bind) get(ctx context.Context, ectx keypattern.ExecContext) (Record, error) {
	rows, err := b.conn.Query(ctx, b.getQuery.String(), b.getQuery.Args(ectx)...)
	if err != nil {
		return nil, err
	}
	record := make(Record, 10)
	err = pgxscan.ScanOne(&record, rows)
	return record, err
}

func (b *Bind) list(ctx context.Context, ectx keypattern.ExecContext) ([]Record, error) {
	res := make([]Record, 0, 10)
	err := pgxscan.Select(ctx, b.conn, &res, b.listQuery.String(), b.listQuery.Args(ectx)...)
	return res, err
}

func (b *Bind) upsert(ctx context.Context, ectx keypattern.ExecContext, value []byte) error {
	if b.upsertQuery == nil {
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
	_, err := b.conn.Exec(ctx, b.upsertQuery.String(), b.upsertQuery.Args(ectx)...)
	return err
}

func (b *Bind) del(ctx context.Context, ectx keypattern.ExecContext) error {
	if b.delQuery == nil {
		return storage.ErrReadOnly
	}
	_, err := b.conn.Exec(ctx, b.delQuery.String(), b.delQuery.Args(ectx)...)
	return err
}
