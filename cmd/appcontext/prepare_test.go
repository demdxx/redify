package appcontext

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareConfig(t *testing.T) {
	os.Setenv("CACHE_CONNECT", "cache_connect")

	os.Setenv("SOURCE1_CONNECT", "source1_connect")
	os.Setenv("SOURCE1_NOTIFY_CHANNEL", "source1_notify_channel")
	os.Setenv("SOURCE1_BIND1_TABLE_NAME", "source1_bind1_table_name")
	os.Setenv("SOURCE1_BIND1_KEY", "source1_bind1_key")
	os.Setenv("SOURCE1_BIND1_WHERE_EXT", "source1_bind1_where_ext")
	os.Setenv("SOURCE1_BIND1_GET_QUERY", "source1_bind1_get_query")
	os.Setenv("SOURCE1_BIND1_LIST_QUERY", "source1_bind1_list_query")
	os.Setenv("SOURCE1_BIND1_UPSERT_QUERY", "source1_bind1_upsert_query")
	os.Setenv("SOURCE1_BIND1_DEL_QUERY", "source1_bind1_del_query")

	os.Setenv("SOURCE1_BIND1_DATATYPE_MAPPING1_NAME", "source1_bind1_datatype_mapping1_name")
	os.Setenv("SOURCE1_BIND1_DATATYPE_MAPPING1_TYPE", "source1_bind1_datatype_mapping1_type")

	conf := ConfigType{
		Cache: cacheConfig{
			Connect: "${{env.CACHE_CONNECT}}",
		},
		Sources: []dataSource{
			{
				Connect:       "${{env.SOURCE1_CONNECT}}",
				NotifyChannel: "${{env.SOURCE1_NOTIFY_CHANNEL}}",
				Binds: []dataSourceKeyBind{
					{
						TableName:   "${{env.SOURCE1_BIND1_TABLE_NAME}}",
						Key:         "${{env.SOURCE1_BIND1_KEY}}",
						WhereExt:    "${{env.SOURCE1_BIND1_WHERE_EXT}}",
						GetQuery:    "${{env.SOURCE1_BIND1_GET_QUERY}}",
						ListQuery:   "${{env.SOURCE1_BIND1_LIST_QUERY}}",
						UpsertQuery: "${{env.SOURCE1_BIND1_UPSERT_QUERY}}",
						DelQuery:    "${{env.SOURCE1_BIND1_DEL_QUERY}}",
						DatatypeMapping: []DatatypeMapper{
							{
								Name: "${{env.SOURCE1_BIND1_DATATYPE_MAPPING1_NAME}}",
								Type: "${{env.SOURCE1_BIND1_DATATYPE_MAPPING1_TYPE}}",
							},
						},
					},
				},
			},
		},
	}
	conf.Prepare()

	assert.Equal(t, "cache_connect", conf.Cache.Connect)
	assert.Equal(t, "source1_connect", conf.Sources[0].Connect)
	assert.Equal(t, "source1_notify_channel", conf.Sources[0].NotifyChannel)
	assert.Equal(t, "source1_bind1_table_name", conf.Sources[0].Binds[0].TableName)
	assert.Equal(t, "source1_bind1_key", conf.Sources[0].Binds[0].Key)
	assert.Equal(t, "source1_bind1_where_ext", conf.Sources[0].Binds[0].WhereExt)
	assert.Equal(t, "source1_bind1_get_query", conf.Sources[0].Binds[0].GetQuery)
	assert.Equal(t, "source1_bind1_list_query", conf.Sources[0].Binds[0].ListQuery)
	assert.Equal(t, "source1_bind1_upsert_query", conf.Sources[0].Binds[0].UpsertQuery)
	assert.Equal(t, "source1_bind1_del_query", conf.Sources[0].Binds[0].DelQuery)
	assert.Equal(t, "source1_bind1_datatype_mapping1_name", conf.Sources[0].Binds[0].DatatypeMapping[0].Name)
	assert.Equal(t, "source1_bind1_datatype_mapping1_type", conf.Sources[0].Binds[0].DatatypeMapping[0].Type)
}

func TestPrepareItem(t *testing.T) {
	os.Setenv("VAR1", "var1")
	os.Setenv("VAR2", "var2")
	cases := []struct {
		in  string
		out string
	}{
		{
			in:  "test",
			out: "test",
		},
		{
			in:  "${{env.test}}",
			out: "",
		},
		{
			in:  "${{env.test}}test",
			out: "test",
		},
		{
			in:  "test_${{env.VAR1}}",
			out: "test_var1",
		},
		{
			in:  "test_${{env.VAR1 }}_${{ env.VAR2}}",
			out: "test_var1_var2",
		},
	}
	for _, c := range cases {
		out := prepareItem(c.in)
		if out != c.out {
			t.Errorf("prepareItem(%q) == %q, want %q", c.in, out, c.out)
		}
	}
}
