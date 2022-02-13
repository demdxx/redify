//go:build redispub || allstreams
// +build redispub allstreams

package connect

import (
	"github.com/demdxx/redify/internal/storage/pubstream"
)

func init() {
	connectors["redispub"] = pubstream.Open
}
