//go:build nats || allstreams
// +build nats allstreams

package connect

import (
	"github.com/demdxx/redify/internal/storage/pubstream"
)

func init() {
	connectors["nats"] = pubstream.Open
}
