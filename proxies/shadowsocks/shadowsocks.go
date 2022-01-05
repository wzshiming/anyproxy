package shadowsocks

import (
	"context"
	"fmt"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/shadowsocks"
)

func NewServeConn(ctx context.Context, scheme string, address string, conf *anyproxy.Config) (anyproxy.ServeConn, []string, error) {
	if len(conf.Users) != 1 {
		return nil, nil, fmt.Errorf("shadowsocks only supports a single authentication method")
	}
	s, err := shadowsocks.NewSimpleServer(scheme + "://" + conf.Users[0].String() + "@" + address)
	if err != nil {
		return nil, nil, err
	}
	s.Context = ctx
	s.Logger = conf.Logger
	if conf.Dialer != nil {
		s.ProxyDial = conf.Dialer.DialContext
	}
	s.BytesPool = conf.BytesPool
	return s, nil, nil
}
