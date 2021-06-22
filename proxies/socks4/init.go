package socks4

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("socks4", NewServeConn)
	anyproxy.Register("socks4a", NewServeConn)
}
