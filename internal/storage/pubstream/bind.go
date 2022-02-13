package pubstream

import (
	"context"
	"sync/atomic"

	nc "github.com/geniusrabbit/notificationcenter"
)

type bind struct {
	counter uint64
	dbnum   int
	key     string
}

func (b *bind) Publish(ctx context.Context, pub nc.Publisher, value []byte) error {
	_ = atomic.AddUint64(&b.counter, 1)
	return pub.Publish(ctx, value)
}

func (b *bind) CountPubs() uint64 {
	return atomic.LoadUint64(&b.counter)
}
