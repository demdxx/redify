package connect

import (
	"context"
	"net/url"

	"github.com/demdxx/redify/internal/storage"
	"github.com/pkg/errors"
)

type connector func(ctx context.Context, connURL string) (storage.Driver, error)

var (
	connectors         = map[string]connector{}
	ErrUndifinedDriver = errors.New("undefined driver")
)

func Connect(ctx context.Context, connURL string) (storage.Driver, error) {
	u, err := url.Parse(connURL)
	if err != nil {
		return nil, err
	}
	if cfnk := connectors[u.Scheme]; cfnk != nil {
		return cfnk(ctx, connURL)
	}
	return nil, errors.Wrap(ErrUndifinedDriver, u.Scheme)
}
