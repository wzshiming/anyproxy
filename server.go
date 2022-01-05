package anyproxy

import (
	"context"
	"fmt"
	"net"
	"net/url"

	_ "github.com/wzshiming/shadowsocks/init"
)

// BytesPool is an interface for getting and returning temporary
// bytes for use by io.CopyBuffer.
type BytesPool interface {
	Get() []byte
	Put([]byte)
}

var schemeMap = map[string]SchemeFunc{}

func Register(scheme string, fun SchemeFunc) {
	schemeMap[scheme] = fun
}

type Config struct {
	Users     []*url.Userinfo
	Dialer    Dialer
	Logger    Logger
	BytesPool BytesPool
}

type SchemeFunc func(ctx context.Context, scheme string, address string, conf *Config) (ServeConn, []string, error)

type ServeConn interface {
	ServeConn(conn net.Conn)
}

type proxyURLs interface {
	ProxyURLs() []string
}

type proxyURL interface {
	ProxyURL() string
}

func NewServeConn(ctx context.Context, scheme string, address string, conf *Config) (ServeConn, []string, error) {
	sch, ok := schemeMap[scheme]
	if !ok || sch == nil {
		return nil, nil, fmt.Errorf("can't support scheme %q", scheme)
	}
	return sch(ctx, scheme, address, conf)
}
