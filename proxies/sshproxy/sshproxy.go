package sshproxy

import (
	"context"
	"fmt"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/sshproxy"
	"golang.org/x/crypto/ssh"
)

var patterns = pattern.Pattern[pattern.SSH]

func NewServeConn(ctx context.Context, scheme string, address string, conf *anyproxy.Config) (anyproxy.ServeConn, []string, error) {
	s, err := sshproxy.NewSimpleServer(scheme + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if conf.Users != nil {
		auth := map[string]string{}
		for _, user := range conf.Users {
			password, _ := user.Password()
			auth[user.Username()] = password
		}
		s.ServerConfig.PasswordCallback = func(conn ssh.ConnMetadata, pwd []byte) (*ssh.Permissions, error) {
			if p, ok := auth[conn.User()]; ok && p == string(pwd) {
				return nil, nil
			}
			return nil, fmt.Errorf("denied")
		}
		s.ServerConfig.NoClientAuth = false
	}
	s.Context = ctx
	s.Logger = conf.Logger
	if conf.Dialer != nil {
		s.ProxyDial = conf.Dialer.DialContext
	}
	if conf.ListenConfig != nil {
		s.ProxyListen = conf.ListenConfig.Listen
	}
	s.BytesPool = conf.BytesPool
	return s, patterns, nil
}
