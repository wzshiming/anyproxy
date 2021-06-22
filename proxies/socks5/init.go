package socks5

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("socks5", NewServeConn)
	anyproxy.Register("socks5h", NewServeConn)
}
