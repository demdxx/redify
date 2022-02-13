//go:build redispub || allstreams
// +build redispub allstreams

package pubstream

import (
	"context"

	nc "github.com/geniusrabbit/notificationcenter"
	"github.com/geniusrabbit/notificationcenter/redis"
)

func init() {
	publisherConnectors["redispub"] = func(ctx context.Context, url string) (nc.Publisher, error) {
		return redis.NewPublisher(redis.WithRedisURL(url))
	}
}
