//go:build kafka || allstreams
// +build kafka allstreams

package connect

import (
	"github.com/demdxx/redify/internal/storage/pubstream"
)

func init() {
	connectors["kafka"] = pubstream.Open
}
