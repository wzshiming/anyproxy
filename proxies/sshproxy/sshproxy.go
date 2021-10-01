package sshproxy

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
	"github.com/wzshiming/sshproxy"
	"golang.org/x/crypto/ssh"
)

var patterns = pattern.Pattern[pattern.SSH]

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger anyproxy.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	s, err := sshproxy.NewSimpleServer(sch + "://" + address)
	if err != nil {
		return nil, nil, err
	}
	if users != nil {
		auth := map[string]string{}
		for _, user := range users {
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
	s.Logger = logger
	s.ProxyDial = dial.DialContext
	s.BytesPool = pool
	return s, patterns, nil
}
