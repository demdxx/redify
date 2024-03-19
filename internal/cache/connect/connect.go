package connect

import (
	"fmt"
	"strings"
	"time"

	"github.com/demdxx/redify/internal/cache"
	"github.com/demdxx/redify/internal/cache/rediscache"
	"github.com/demdxx/redify/internal/cache/simplecache"
)

func Connect(connect string, size int, ttl time.Duration) (cache.Cacher, error) {
	switch {
	case strings.HasPrefix(connect, "redis://"):
		return rediscache.New(connect, ttl)
	case connect == "memory":
		return simplecache.New(size, int(ttl.Seconds()))
	default:
		return nil, fmt.Errorf("invalid cache connect: %s", connect)
	}
}
