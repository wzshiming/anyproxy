package httpproxy

import (
	"context"
	"log"
	"net"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/anyproxy/internal/warpping"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/httpproxy"
)

var patterns = append(pattern.Pattern[pattern.HTTP], pattern.Pattern[pattern.HTTP2]...)

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger *log.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	s, err := httpproxy.NewSimpleServer(sch + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	s.Server.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}
	if users != nil {
		auth := map[string]string{}
		for _, user := range users {
			password, _ := user.Password()
			auth[user.Username()] = password
		}
		s.Authentication = httpproxy.BasicAuthFunc(func(username, password string) bool {
			return auth[username] == password
		})
	}
	s.Server.ErrorLog = logger
	s.ProxyDial = dial.DialContext
	s.BytesPool = pool

	return warpping.NewWarpHttpConn(&s.Server), patterns, nil
}
