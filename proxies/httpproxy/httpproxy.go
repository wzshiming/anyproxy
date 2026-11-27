package httpproxy

import (
	"context"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/httpproxy"
)

var patterns = append(pattern.Pattern[pattern.HTTP], pattern.Pattern[pattern.HTTP2]...)

func NewServeConn(ctx context.Context, scheme string, address string, conf *anyproxy.Config) (anyproxy.ServeConn, []string, error) {
	s, err := httpproxy.NewSimpleServer(scheme + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	s.Context = ctx
	if conf.Users != nil {
		auth := map[string]string{}
		for _, user := range conf.Users {
			password, _ := user.Password()
			auth[user.Username()] = password
		}
		s.Authentication = httpproxy.BasicAuthFunc(func(username, password string) bool {
			expected, ok := auth[username]
			return ok && expected == password
		})
	}
	s.Logger = conf.Logger
	if conf.Dialer != nil {
		s.ProxyDial = conf.Dialer.DialContext
	}
	s.BytesPool = conf.BytesPool
	return &s.Server, patterns, nil
}
