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
	"github.com/wzshiming/socks4"
	"github.com/wzshiming/socks5"
)

// BytesPool is an interface for getting and returning temporary
// bytes for use by io.CopyBuffer.
type BytesPool interface {
	Get() []byte
	Put([]byte)
}

func newServer(ctx context.Context, addr string, dial Dialer, logger *log.Logger, pool BytesPool) (*warpPatternServer, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http":
		s, err := httpproxy.NewSimpleServer(addr)
		if err != nil {
			return nil, err
		}
		s.Server.BaseContext = func(listener net.Listener) context.Context {
			return ctx
		}
		s.Server.ErrorLog = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return newWarpPatternServer(newWarpHttpProxySimpleServer(s), []string{pattern.HTTP, pattern.HTTP2}), nil
	case "socks4", "socks4a":
		s, err := socks4.NewSimpleServer(addr)
		if err != nil {
			return nil, err
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return newWarpPatternServer(s, []string{pattern.SOCKS4}), nil
	case "socks5", "socks5h":
		s, err := socks5.NewSimpleServer(addr)
		if err != nil {
			return nil, err
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return newWarpPatternServer(s, []string{pattern.SOCKS5}), nil
	case "ss", "shadowsocks":
		s, err := shadowsocks.NewSimpleServer(addr)
		if err != nil {
			return nil, err
		}
		s.Context = ctx
		s.Logger = logger
		s.ProxyDial = dial.DialContext
		s.BytesPool = pool
		return newWarpPatternServer(s, nil), nil
	}
	return nil, fmt.Errorf("unsupported protocol '%s'", u.Scheme)
}

type serveConn interface {
	ServeConn(conn net.Conn)
	ProxyURL() string
}

type warpPatternServer struct {
	serveConn
	patterns []string
}

func (p *warpPatternServer) Patterns() []string {
	return p.patterns
}

func newWarpPatternServer(s serveConn, p []string) *warpPatternServer {
	return &warpPatternServer{serveConn: s, patterns: p}
}

type warpHttpProxySimpleServer struct {
	*httpproxy.SimpleServer
	warpHttpConn
}

func newWarpHttpProxySimpleServer(s *httpproxy.SimpleServer) serveConn {
	return warpHttpProxySimpleServer{
		SimpleServer: s,
		warpHttpConn: warpHttpConn{&s.Server},
	}
}
