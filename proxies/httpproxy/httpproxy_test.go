package httpproxy

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wzshiming/anyproxy"
)

type countingDialer struct{ calls atomic.Int32 }

func (dialer *countingDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	dialer.calls.Add(1)
	return nil, errors.New("dial blocked by test")
}

func TestBasicAuthRejectsUnknownUsers(t *testing.T) {
	users := []*url.Userinfo{url.UserPassword("alice", "secret"), url.UserPassword("bob", "")}
	cases := []struct {
		name   string
		users  []*url.Userinfo
		auth   string
		denied bool
	}{
		{"unknown user with empty password", users, "ghost:", true},
		{"known user with wrong password", users, "alice:wrong", true},
		{"known user with valid password", users, "alice:secret", false},
		{"known user with configured empty password", users, "bob:", false},
		{"no users configured", nil, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dialer := &countingDialer{}
			serveConn, _, err := NewServeConn(t.Context(), "http", "127.0.0.1:0", &anyproxy.Config{Users: tc.users, Dialer: dialer})
			if err != nil {
				t.Fatal(err)
			}
			clientConn, serverConn := net.Pipe()
			clientConn.SetDeadline(time.Now().Add(5 * time.Second))
			done := make(chan struct{})
			go func() {
				defer close(done)
				serveConn.ServeConn(serverConn)
			}()
			defer func() {
				clientConn.Close()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Error("ServeConn did not return after the client closed")
					return
				}
				wantDials := int32(1)
				if tc.denied {
					wantDials = 0
				}
				if got := dialer.calls.Load(); got != wantDials {
					t.Errorf("dial calls = %d, want %d", got, wantDials)
				}
			}()

			req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.auth != "" {
				req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(tc.auth)))
			}
			if err := req.WriteProxy(clientConn); err != nil {
				t.Fatal(err)
			}
			resp, err := http.ReadResponse(bufio.NewReader(clientConn), req)
			if err != nil {
				t.Fatal(err)
			}
			// An accepted request reaches the failing dialer, so anything but 407 means accepted.
			if tc.denied != (resp.StatusCode == http.StatusProxyAuthRequired) {
				t.Errorf("status = %d, want denied = %v", resp.StatusCode, tc.denied)
			}
		})
	}
}
