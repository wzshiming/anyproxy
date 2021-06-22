package pprof

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("pprof", NewServeConn)
}
