package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereStmt(t *testing.T) {
	wr := WhereStmt{`col1`: `val1`, `col2`: `val2`}
	wrs := wr.Where(`"`, "ext")
	res := false ||
		wrs == ` WHERE "col1"=val1 AND "col2"=val2 AND ext` ||
		wrs == ` WHERE "col2"=val2 AND "col1"=val1 AND ext`
	assert.True(t, res)
}

func TestSQLData(t *testing.T) {
	wr := DataFields{`col1`: `val1`, `col2`: `val2`}
	assert.True(t, false ||
		wr.Columns(`"`) == `"col1", "col2"` ||
		wr.Columns(`"`) == `"col2", "col1"`)
	assert.True(t, false ||
		wr.Values() == `val1, val2` ||
		wr.Values() == `val2, val1`)
	assert.True(t, false ||
		wr.SetValues(`"`) == `"col1"=val1, "col2"=val2` ||
		wr.SetValues(`"`) == `"col2"=val2, "col1"=val1`)
}
