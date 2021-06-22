package anyproxy

import (
	"context"
	"fmt"
	"log"
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

type SchemeFunc func(ctx context.Context, sch, address string, users []*url.Userinfo, dial Dialer, logger *log.Logger, pool BytesPool) (ServeConn, []string, error)

type ServeConn interface {
	ServeConn(conn net.Conn)
}

type proxyURLs interface {
	ProxyURLs() []string
}

type proxyURL interface {
	ProxyURL() string
}

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial Dialer, logger *log.Logger, pool BytesPool) (ServeConn, []string, error) {
	scheme, ok := schemeMap[sch]
	if !ok || scheme == nil {
		return nil, nil, fmt.Errorf("can't support scheme %q", sch)
	}
	return scheme(ctx, sch, address, users, dial, logger, pool)
}
