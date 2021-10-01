package sshproxy

import (
	"github.com/wzshiming/anyproxy"
)

func init() {
	anyproxy.Register("ssh", NewServeConn)
}
