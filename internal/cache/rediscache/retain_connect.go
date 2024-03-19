package rediscache

import (
	"io"
	"sync"

	"github.com/go-redis/redis/v8"
)

type retainConnect struct {
	mx sync.Mutex
	redis.Cmdable
	refs int
}

func (rc *retainConnect) inc() *retainConnect {
	rc.mx.Lock()
	defer rc.mx.Unlock()
	rc.refs++
	return rc
}

func (rc *retainConnect) Close() error {
	rc.mx.Lock()
	defer rc.mx.Unlock()
	if rc.refs--; rc.Cmdable != nil && rc.refs < 0 {
		var err error
		if cl := rc.Cmdable.(io.Closer); cl != nil {
			err = cl.Close()
		}
		rc.Cmdable = nil
		return err
	}
	return nil
}
