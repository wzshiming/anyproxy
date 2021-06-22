package shadowsocks

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/shadowsocks"
)

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger *log.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	if len(users) != 1 {
		return nil, nil, fmt.Errorf("shadowsocks only supports a single authentication method")
	}
	s, err := shadowsocks.NewSimpleServer(sch + "://" + users[0].String() + "@" + address)
	if err != nil {
		return nil, nil, err
	}
	s.Context = ctx
	s.Logger = logger
	s.ProxyDial = dial.DialContext
	s.BytesPool = pool
	return s, nil, nil
}
