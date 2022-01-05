package socks4

import (
	"context"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/socks4"
)

var patterns = pattern.Pattern[pattern.SOCKS4]

func NewServeConn(ctx context.Context, scheme string, address string, conf *anyproxy.Config) (anyproxy.ServeConn, []string, error) {
	s, err := socks4.NewSimpleServer(scheme + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if conf.Users != nil {
		auth := map[string]struct{}{}
		for _, user := range conf.Users {
			auth[user.Username()] = struct{}{}
		}
		s.Authentication = socks4.AuthenticationFunc(func(cmd socks4.Command, username string) bool {
			_, ok := auth[username]
			return ok
		})
	}
	s.Context = ctx
	s.Logger = conf.Logger
	if conf.Dialer != nil {
		s.ProxyDial = conf.Dialer.DialContext
	}
	s.BytesPool = conf.BytesPool
	return s, patterns, nil
}
