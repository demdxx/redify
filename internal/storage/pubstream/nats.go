//go:build nats || allstreams
// +build nats allstreams

package pubstream

import (
	"context"

	nc "github.com/geniusrabbit/notificationcenter"
	"github.com/geniusrabbit/notificationcenter/nats"
)

func init() {
	publisherConnectors["nats"] = func(ctx context.Context, url string) (nc.Publisher, error) {
		return nats.NewPublisher(nats.WithNatsURL(url))
	}
}
