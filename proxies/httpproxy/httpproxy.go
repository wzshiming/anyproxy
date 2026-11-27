package httpproxy

import (
	"context"
	"net/http"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/httpproxy"
)

// patterns matches the requests of HTTP/1.x proxy only, which is CONNECT
// and the request with an absolute URL of http.
var patterns = proxyPatterns()

func proxyPatterns() []string {
	methods := pattern.Pattern[pattern.HTTP]
	patterns := make([]string, 0, len(methods))
	for _, method := range methods {
		if method == http.MethodConnect+" " {
			patterns = append(patterns, method)
			continue
		}
		patterns = append(patterns, method+"http://")
	}
	return patterns
}

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
