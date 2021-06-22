package anyproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/httpproxy"
	"github.com/wzshiming/shadowsocks"
	_ "github.com/wzshiming/shadowsocks/init"
	"github.com/wzshiming/socks4"
	"github.com/wzshiming/socks5"
)

// BytesPool is an interface for getting and returning temporary
// bytes for use by io.CopyBuffer.
type BytesPool interface {
	Get() []byte
	Put([]byte)
}

type scheme int

const (
	_ scheme = iota
	schemeHTTP
	schemeSocks4
	schemeSocks5
	schemeShadowsocks
)

var schemeMap = map[string]scheme{
	"http":        schemeHTTP,
	"socks4":      schemeSocks4,
	"socks4a":     schemeSocks4,
	"socks5":      schemeSocks5,
	"socks5h":     schemeSocks5,
	"shadowsocks": schemeShadowsocks,
	"ss":          schemeShadowsocks,
}

func newServeConn(ctx context.Context, scheme scheme, sch, address string, users []*url.Userinfo, dial Dialer, logger *log.Logger, pool BytesPool) (ServeConn, []string, error) {
	switch scheme {
	case schemeHTTP:
		s, err := httpproxy.NewSimpleServer(sch + "://" + address)
		if err != nil {
			return nil, nil, err
		}
		s.Server.BaseContext = func(listener net.Listener) context.Context {
			return ctx
		}
		if users != nil {
			auth := map[string]string{}
			for _, user := range users {
				password, _ := user.Password()
				auth[user.Username()] = password
			}
			s.Authentication = httpproxy.BasicAuthFunc(func(username, password string) bool {
				return auth[username] == password
			})
		}
		s.Server.ErrorLog = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return newWarpHttpProxySimpleServer(s), []string{pattern.HTTP, pattern.HTTP2}, nil
	case schemeSocks4:
		s, err := socks4.NewSimpleServer(sch + "://" + address)
		if err != nil {
			return nil, nil, err
		}
		if users != nil {
			auth := map[string]struct{}{}
			for _, user := range users {
				auth[user.Username()] = struct{}{}
			}
			s.Authentication = socks4.AuthenticationFunc(func(cmd socks4.Command, username string) bool {
				_, ok := auth[username]
				return ok
			})
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return s, []string{pattern.SOCKS4}, nil
	case schemeSocks5:
		s, err := socks5.NewSimpleServer(sch + "://" + address)
		if err != nil {
			return nil, nil, err
		}
		if users != nil {
			auth := map[string]string{}
			for _, user := range users {
				password, _ := user.Password()
				auth[user.Username()] = password
			}
			s.Authentication = socks5.AuthenticationFunc(func(cmd socks5.Command, username, password string) bool {
				return auth[username] == password
			})
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return s, []string{pattern.SOCKS5}, nil
	case schemeShadowsocks:
		if len(users) != 1 {
			return nil, nil, fmt.Errorf("shadowsocks only supports a single authentication method")
		}
		s, err := shadowsocks.NewSimpleServer(sch + "://" + address)
		if err != nil {
			return nil, nil, err
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return s, nil, nil
	}
	return nil, nil, fmt.Errorf("unsupported protocol %q", scheme)
}

type ServeConn interface {
	ServeConn(conn net.Conn)
}

type proxyURLs interface {
	ProxyURLs() []string
}

type proxyURL interface {
	ProxyURL() string
}
