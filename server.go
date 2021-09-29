package anyproxy

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/wzshiming/cmux"
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

type SchemeFunc func(ctx context.Context, sch, address string, users []*url.Userinfo, dial Dialer, logger Logger, pool BytesPool) (ServeConn, []string, error)

type ServeConn interface {
	ServeConn(conn net.Conn)
}

type proxyURLs interface {
	ProxyURLs() []string
}

type proxyURL interface {
	ProxyURL() string
}

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial Dialer, logger Logger, pool BytesPool) (ServeConn, []string, error) {
	scheme, ok := schemeMap[sch]
	if !ok || scheme == nil {
		return nil, nil, fmt.Errorf("can't support scheme %q", sch)
	}
	return scheme(ctx, sch, address, users, dial, logger, pool)
}

func NewServeConnWithAllScheme(ctx context.Context, address string, users []*url.Userinfo, dial Dialer, logger Logger, pool BytesPool) (ServeConn, []string, error) {
	c := cmux.NewCMux()
	for _, sch := range ListScheme() {
		s, patterns, err := NewServeConn(ctx, sch, address, users, dial, logger, pool)
		if patterns == nil {
			err = c.NotFound(s)
			if err != nil {
				return nil, nil, err
			}
		} else {
			err = c.HandlePrefix(s, patterns...)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return c, nil, nil
}

func ListScheme() []string {
	schemes := make([]string, 0, len(schemeMap))
	for sch := range schemeMap {
		schemes = append(schemes, sch)
	}
	return schemes
}
