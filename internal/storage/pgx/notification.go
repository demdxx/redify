package pgx

import (
	"encoding/json"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/redify/internal/keypattern"
)

type Notification struct {
	Table string          `json:"table"`
	Data  json.RawMessage `json:"data"`
}

func (n *Notification) unmarshal(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *Notification) ectx() (e keypattern.ExecContext, err error) {
	var m map[string]any
	err = json.Unmarshal(n.Data, &m)
	if err == nil {
		e = make(keypattern.ExecContext, len(m))
		for k, v := range m {
			e[k] = gocast.Str(v)
		}
	}
	return e, err
}
