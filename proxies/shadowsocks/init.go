package shadowsocks

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("shadowsocks", NewServeConn)
	anyproxy.Register("ss", NewServeConn)
}
