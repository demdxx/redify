package sql

import (
	"context"
	"path/filepath"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
)

type (
	Record = storage.Record
	query  = Query
)

type Syntax interface {
	UpsertQuery(tableName string, insertFields DataFields, keyFields []string) string
	GetQuery(tableName string, where WhereStmt, whereExt string) string
	SelectQuery(tableName string, where WhereStmt, whereExt string) string
	DeleteQuery(tableName string, where WhereStmt, whereExt string) string
}

type BindAbstract struct {
	DBNum            int
	Pattern          *keypattern.Pattern
	Syntax           Syntax
	GetQuery         *query
	ListQuery        *query
	UpsertQuery      *query
	DelQuery         *query
	DatatypesMapping []storage.DatatypeMapper
}

func NewBindAbstract(dbnum int, syntax Syntax, pattern, getQuery, listQuery, upsertQuery, delQuery string, datatypesMapping []storage.DatatypeMapper) *BindAbstract {
	return &BindAbstract{
		DBNum:            dbnum,
		Pattern:          keypattern.NewPatternFromExpression(pattern),
		Syntax:           syntax,
		GetQuery:         ParseQuery(getQuery),
		ListQuery:        ParseQuery(listQuery),
		UpsertQuery:      ParseQuery(upsertQuery),
		DelQuery:         ParseQuery(delQuery),
		DatatypesMapping: datatypesMapping,
	}
}

func NewBindAbstractFromTableName(dbnum int, syntax Syntax, pattern, tableName, whereExt string, datatypesMapping []storage.DatatypeMapper, readonly bool) *BindAbstract {
	var (
		ptrObj           = keypattern.NewPatternFromExpression(pattern)
		keyFields        = ptrObj.Keys()
		whereConds       = make(WhereStmt, len(keyFields))
		dataValues       = make(DataFields, len(keyFields))
		delQyeryObj      *query
		upinsertQyeryObj *query
	)
	if keys := ptrObj.Keys(); len(keys) > 0 {
		for _, key := range keys {
			whereConds[key] = "{{" + key + "}}"
			dataValues[key] = "{{" + key + "}}"
		}
	}
	if !readonly {
		delQyeryObj = ParseQuery(syntax.DeleteQuery(tableName, whereConds, whereExt))
		upinsertQyeryObj = ParseQuery(syntax.UpsertQuery(tableName, dataValues, keyFields))
	}
	return &BindAbstract{
		DBNum:            dbnum,
		Syntax:           syntax,
		Pattern:          ptrObj,
		GetQuery:         ParseQuery(syntax.GetQuery(tableName, whereConds, whereExt)),
		ListQuery:        ParseQuery(syntax.SelectQuery(tableName, nil, "")),
		DelQuery:         delQyeryObj,
		UpsertQuery:      upinsertQyeryObj,
		DatatypesMapping: datatypesMapping,
	}
}

func (b *BindAbstract) TableName() string {
	return b.GetQuery.TableName
}

func (b *BindAbstract) MatchKey(key string, ectx keypattern.ExecContext) bool {
	return b.Pattern.Match(key, ectx)
}

func (b *BindAbstract) MatchPattern(pt string, ectx keypattern.ExecContext) bool {
	ok, _ := filepath.Match(pt, b.Pattern.String())
	return ok
}

func (b *BindAbstract) Get(ctx context.Context, ectx keypattern.ExecContext) (Record, error) {
	return nil, nil
}

func (b *BindAbstract) List(ctx context.Context, ectx keypattern.ExecContext) ([]Record, error) {
	return nil, nil
}

func (b *BindAbstract) Upsert(ctx context.Context, ectx keypattern.ExecContext, value []byte) error {
	return nil
}

func (b *BindAbstract) Del(ctx context.Context, ectx keypattern.ExecContext) error {
	return nil
}
