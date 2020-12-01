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

func NewAnyProxy(dial Dialer, logger *log.Logger) *AnyProxy {
	httpd := http.Server{
		ErrorLog: logger,
		Handler: &httpproxy.ProxyHandler{
			Logger:    logger,
			ProxyDial: dial.DialContext,
		},
	}
	socks4d := socks4.Server{
		Logger:    logger,
		ProxyDial: dial.DialContext,
	}
	socks5d := socks5.Server{
		Logger:    logger,
		ProxyDial: dial.DialContext,
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
