package socks5

import (
	"context"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/socks5"
)

var patterns = pattern.Pattern[pattern.SOCKS5]

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger anyproxy.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	s, err := socks5.NewSimpleServer(sch + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if users != nil {
		auth := map[string]string{}
		for _, user := range users {
			password, _ := user.Password()
			auth[user.Username()] = password
		}
		s.Authentication = socks5.AuthenticationFunc(func(cmd socks5.Command, username, password string) bool {
			return auth[username] == password
		})
	}
	s.Context = ctx
	s.Logger = logger
	s.ProxyDial = dial.DialContext
	s.BytesPool = pool
	return s, patterns, nil
}
