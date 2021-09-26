package socks4

import (
	"context"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/socks4"
)

var patterns = pattern.Pattern[pattern.SOCKS4]

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger anyproxy.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	s, err := socks4.NewSimpleServer(sch + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if users != nil {
		auth := map[string]struct{}{}
		for _, user := range users {
			auth[user.Username()] = struct{}{}
		}
		s.Authentication = socks4.AuthenticationFunc(func(cmd socks4.Command, username string) bool {
			_, ok := auth[username]
			return ok
		})
	}
	s.Context = ctx
	s.Logger = logger
	s.ProxyDial = dial.DialContext
	s.BytesPool = pool
	return s, patterns, nil
}
