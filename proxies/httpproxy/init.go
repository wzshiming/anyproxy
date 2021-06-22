package httpproxy

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("http", NewServeConn)
}
