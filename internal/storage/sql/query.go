package sql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/demdxx/redify/internal/keypattern"
)

var (
	reTableSelect = regexp.MustCompile(`(?mi)select\s+.*\s+from\s+([^\s]+)`)
	reTableInsert = regexp.MustCompile(`(?mi)insert\s+into\s+([^\s]+)`)
	reTableDelete = regexp.MustCompile(`(?mi)delete\s+from\s+([^\s]+)`)
)

type Query struct {
	queryStr  string
	TableName string
	arguments []string // List of arguments in the correct order
}

func ParseQuery(q string) *Query {
	if q == "" {
		return nil
	}
	var (
		vars      = keypattern.ExtractVars(q)
		args      = make([]string, 0, len(vars))
		tableName string
		r         []string
	)
	for i, v := range vars {
		args = append(args, v[1])
		q = strings.ReplaceAll(q, v[0], fmt.Sprintf("$%d", i+1))
	}
	q2 := strings.TrimSpace(strings.ToLower(q))
	switch {
	case strings.HasPrefix(q2, "select"):
		r = reTableSelect.FindStringSubmatch(q)
	case strings.HasPrefix(q2, "insert"):
		r = reTableInsert.FindStringSubmatch(q)
	case strings.HasPrefix(q2, "delete"):
		r = reTableDelete.FindStringSubmatch(q)
	}
	if len(r) > 0 {
		tableName = r[1]
	}
	return &Query{queryStr: q, TableName: tableName, arguments: args}
}

func (q *Query) String() string {
	return q.queryStr
}

func (q *Query) Args(ectx keypattern.ExecContext) []any {
	res := make([]any, 0, len(q.arguments))
	for _, arg := range q.arguments {
		res = append(res, ectx[arg])
	}
	return res
}
