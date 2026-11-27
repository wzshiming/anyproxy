package anyproxy

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
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
	conn = connCompatibilityReadDeadline{conn}
	listener := newSingleConnListener(conn)
	w.Serve(listener)
}

// aLongTimeAgo is the read deadline net/http sets to abort a pending read.
var aLongTimeAgo = time.Unix(1, 0)

// connCompatibilityReadDeadline keeps a wrapped net.Conn working with
// net/http, which aborts the pending read by a deadline in the past.
type connCompatibilityReadDeadline struct {
	net.Conn
}

func (d connCompatibilityReadDeadline) SetReadDeadline(t time.Time) error {
	if t.Equal(aLongTimeAgo) {
		t = time.Now().Add(1 * time.Second)
	}
	return d.Conn.SetReadDeadline(t)
}
