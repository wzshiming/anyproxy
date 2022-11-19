package anyproxy

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/wzshiming/httpproxy"
)

var ErrNetClosing = errors.New("use of closed network connection")

type singleConnListener struct {
	addr net.Addr
	ch   chan net.Conn
	once sync.Once
}

func newSingleConnListener(conn net.Conn) net.Listener {
	ch := make(chan net.Conn, 1)
	ch <- conn
	return &singleConnListener{
		addr: conn.LocalAddr(),
		ch:   ch,
	}
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	conn, ok := <-l.ch
	if !ok || conn == nil {
		return nil, ErrNetClosing
	}
	return &connCloser{
		l:    l,
		Conn: conn,
	}, nil
}

func (l *singleConnListener) shutdown() error {
	l.once.Do(func() {
		close(l.ch)
	})
	return nil
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return l.addr
}

type connCloser struct {
	l *singleConnListener
	net.Conn
}

func (c *connCloser) Close() error {
	c.l.shutdown()
	return c.Conn.Close()
}

type httpServeConn struct {
	*http.Server
}

func NewHttpServeConn(s *http.Server) ServeConn {
	return &httpServeConn{s}
}

func (w httpServeConn) ServeConn(conn net.Conn) {
	listener := newSingleConnListener(conn)
	listener = httpproxy.NewListenerCompatibilityReadDeadline(listener)
	w.Serve(listener)
}
