package anyproxy

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"

	"github.com/wzshiming/cmux"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/httpproxy"
	"github.com/wzshiming/socks4"
	"github.com/wzshiming/socks5"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type AnyProxy struct {
	Socks4 socks4.Server
	Socks5 socks5.Server
	Http   http.Server
	CMux   *cmux.CMux
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type BytesPool httpproxy.BytesPool

func NewAnyProxy(ctx context.Context, dial Dialer, logger *log.Logger, pool BytesPool) *AnyProxy {
	httpd := http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		ErrorLog: logger,
		Handler: h2c.NewHandler(&httpproxy.ProxyHandler{
			Logger:    logger,
			ProxyDial: dial.DialContext,
			BytesPool: pool,
		}, &http2.Server{}),
	}
	socks4d := socks4.Server{
		Context:   ctx,
		Logger:    logger,
		ProxyDial: dial.DialContext,
		BytesPool: pool,
	}
	socks5d := socks5.Server{
		Context:   ctx,
		Logger:    logger,
		ProxyDial: dial.DialContext,
		BytesPool: pool,
	}

	proxy := &AnyProxy{
		Socks4: socks4d,
		Socks5: socks5d,
		Http:   httpd,
		CMux:   cmux.NewCMux(),
	}

	proxy.CMux.HandleRegexp(pattern.Socks4, &proxy.Socks4)
	proxy.CMux.HandleRegexp(pattern.Socks5, &proxy.Socks5)
	proxy.CMux.NotFound(warpHttpConn{&proxy.Http})
	return proxy
}

// ServeConn is used to serve a single connection.
func (s *AnyProxy) ServeConn(conn net.Conn) {
	conn = &connBuffReader{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
	s.CMux.ServeConn(conn)
}

func (s *AnyProxy) ListenAndServe(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
}
