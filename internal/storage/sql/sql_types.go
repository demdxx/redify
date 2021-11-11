package sql

import "bytes"

type (
	WhereStmt  map[string]string
	DataFields map[string]string
)

func (ws WhereStmt) Where(escape, whereExt string) string {
	var buf bytes.Buffer
	if ws != nil && (len(ws) > 0 || whereExt != "") {
		buf.WriteString(` WHERE `)
		for k, v := range ws {
			if buf.Len() > 7 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(escape)
			buf.WriteString(k)
			buf.WriteString(escape)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
		if whereExt != "" {
			if buf.Len() > 7 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(whereExt)
		}
	}
	return buf.String()
}

func (df DataFields) Columns(escape string) string {
	var buf bytes.Buffer
	for k := range df {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(escape)
		buf.WriteString(k)
		buf.WriteString(escape)
	}
	return buf.String()
}

func (df DataFields) Values() string {
	var buf bytes.Buffer
	for _, v := range df {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v)
	}
	return buf.String()
}

func (df DataFields) SetValues(escape string) string {
	var buf bytes.Buffer
	for k, v := range df {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(escape)
		buf.WriteString(k)
		buf.WriteString(escape)
		buf.WriteByte('=')
		buf.WriteString(v)
	}
	return buf.String()
}
