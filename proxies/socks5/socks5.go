package socks5

import (
	"context"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/socks5"
)

var patterns = pattern.Pattern[pattern.SOCKS5]

func NewServeConn(ctx context.Context, scheme string, address string, conf *anyproxy.Config) (anyproxy.ServeConn, []string, error) {
	s, err := socks5.NewSimpleServer(scheme + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if conf.Users != nil {
		auth := map[string]string{}
		for _, user := range conf.Users {
			password, _ := user.Password()
			auth[user.Username()] = password
		}
		s.Authentication = socks5.AuthenticationFunc(func(cmd socks5.Command, username, password string) bool {
			return auth[username] == password
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
