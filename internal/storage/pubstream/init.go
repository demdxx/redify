package pubstream

import (
	"context"
	"net/url"

	nc "github.com/geniusrabbit/notificationcenter/v2"
	"github.com/pkg/errors"
)

// ErrUnsupportedScheme in case if scheme is not defined
var ErrUnsupportedScheme = errors.New(`unsupported scheme`)

type publisherConnector func(ctx context.Context, url string) (nc.Publisher, error)

var publisherConnectors = map[string]publisherConnector{}

// connect stream publisher from URL
func connect(ctx context.Context, urlStr string) (nc.Publisher, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	conn := publisherConnectors[parsedURL.Scheme]
	if conn == nil {
		return nil, errors.Wrap(ErrUnsupportedScheme, parsedURL.Scheme)
	}
	pub, err := conn(ctx, urlStr)
	return pub, err
}
