package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereStmt(t *testing.T) {
	wr := WhereStmt{`col1`: `val1`, `col2`: `val2`}
	assert.Contains(t, []string{
		` WHERE "col1"=val1 AND "col2"=val2 AND ext`,
		` WHERE "col2"=val2 AND "col1"=val1 AND ext`,
	}, wr.Where(`"`, "ext"))
}

func TestSQLData(t *testing.T) {
	wr := DataFields{`col1`: `val1`, `col2`: `val2`}
	assert.Contains(t, []string{`"col1", "col2"`, `"col2", "col1"`}, wr.Columns(`"`))
	assert.Contains(t, []string{`val1, val2`, `val2, val1`}, wr.Values())
	assert.Contains(t, []string{`"col1"=val1, "col2"=val2`, `"col2"=val2, "col1"=val1`}, wr.SetValues(`"`))
}
