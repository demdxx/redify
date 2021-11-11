package sql

import "strings"

type AbstractSyntax struct {
	columnEscape string
}

func NewAbstractSyntax(escape string) *AbstractSyntax {
	return &AbstractSyntax{columnEscape: escape}
}

func (sx *AbstractSyntax) UpsertQuery(tableName string, insertFields DataFields, keyFields []string) string {
	return `INSERT INTO ` + tableName + ` VALUES(` + insertFields.Columns(sx.columnEscape) + `) (` + insertFields.Values() + `)` +
		` ON CONFLICT (` + strings.Join(keyFields, ", ") + `) DO UPDATE SET ` + insertFields.SetValues(sx.columnEscape)
}

func (sx *AbstractSyntax) GetQuery(tableName string, where WhereStmt, whereExt string) string {
	return `SELECT * FROM ` + tableName + where.Where(sx.columnEscape, whereExt) + ` LIMIT 1`
}

func (sx *AbstractSyntax) SelectQuery(tableName string, where WhereStmt, whereExt string) string {
	return `SELECT * FROM ` + tableName + where.Where(sx.columnEscape, whereExt)
}

func (sx *AbstractSyntax) DeleteQuery(tableName string, where WhereStmt, whereExt string) string {
	return `DELETE FROM ` + tableName + where.Where(sx.columnEscape, whereExt)
}

type MysqlSyntax struct {
	AbstractSyntax
}

func NewMysqlSyntax() *MysqlSyntax {
	return &MysqlSyntax{
		AbstractSyntax: AbstractSyntax{columnEscape: "`"},
	}
}

func (sx *MysqlSyntax) UpsertQuery(tableName string, insertFields DataFields, keyFields []string) string {
	return `REPLACE INTO ` + tableName + ` VALUES(` + insertFields.Columns(sx.columnEscape) + `) (` + insertFields.Values() + `)`
}
