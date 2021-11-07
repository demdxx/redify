package sql

import (
	"reflect"
	"testing"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		q               string
		ctx             keypattern.ExecContext
		expectQ         string
		extpectArgs     []string
		expectTableName string
		expectVars      []interface{}
	}{
		{
			q:               "SELECT * FROM data WHERE slug={{slug}}, type={{type}}",
			ctx:             keypattern.ExecContext{"slug": "slug1", "type": "type1"},
			expectQ:         "SELECT * FROM data WHERE slug=$1, type=$2",
			expectTableName: "data",
			extpectArgs:     []string{"slug", "type"},
			expectVars:      []interface{}{"slug1", "type1"},
		},
		{
			q:               "INSERT INTO document (slug, type) VALUES({{slug}}, {{type}})",
			ctx:             keypattern.ExecContext{"slug": "slug1", "type": "type1"},
			expectQ:         "INSERT INTO document (slug, type) VALUES($1, $2)",
			expectTableName: "document",
			extpectArgs:     []string{"slug", "type"},
			expectVars:      []interface{}{"slug1", "type1"},
		},
		{
			q:               "DELETE FROM store WHERE slug={{slug}}, type={{type}}",
			ctx:             keypattern.ExecContext{"slug": "slug1", "type": "type1"},
			expectQ:         "DELETE FROM store WHERE slug=$1, type=$2",
			expectTableName: "store",
			extpectArgs:     []string{"slug", "type"},
			expectVars:      []interface{}{"slug1", "type1"},
		},
	}
	for _, test := range tests {
		var (
			q    = ParseQuery(test.q)
			args = q.Args(test.ctx)
		)
		assert.Equal(t, test.expectQ, q.queryStr, "not correct query prepare")
		assert.True(t, reflect.DeepEqual(test.extpectArgs, q.arguments), "invalid parsed params")
		assert.Equal(t, test.expectTableName, q.TableName, "invalid parsed tablename")
		assert.True(t, reflect.DeepEqual(test.expectVars, args), "invalid param extraction")
	}
}
