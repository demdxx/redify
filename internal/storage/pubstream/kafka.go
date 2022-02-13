//go:build kafka || allstreams
// +build kafka allstreams

package pubstream

import (
	"context"

	nc "github.com/geniusrabbit/notificationcenter"
	"github.com/geniusrabbit/notificationcenter/kafka"
)

func init() {
	publisherConnectors["kafka"] = func(ctx context.Context, url string) (nc.Publisher, error) {
		return kafka.NewPublisher(ctx, kafka.WithKafkaURL(url))
	}
}
