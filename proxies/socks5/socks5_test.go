package socks5

import (
	"context"
	"errors"
	"io"
	"net"
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

func TestUserAuthRejectsUnknownUsers(t *testing.T) {
	users := []*url.Userinfo{url.UserPassword("alice", "secret"), url.UserPassword("bob", "")}
	cases := []struct {
		name     string
		username string
		password string
		denied   bool
	}{
		{"unknown user with empty password", "ghost", "", true},
		{"known user with wrong password", "alice", "wrong", true},
		{"known user with valid password", "alice", "secret", false},
		{"known user with configured empty password", "bob", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dialer := &countingDialer{}
			serveConn, _, err := NewServeConn(t.Context(), "socks5", "127.0.0.1:0", &anyproxy.Config{Users: users, Dialer: dialer})
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

			// net.Pipe writes block until read, so each message is sent only after the previous reply.
			exchange := func(send []byte, replyLen int) []byte {
				if _, err := clientConn.Write(send); err != nil {
					t.Fatal(err)
				}
				reply := make([]byte, replyLen)
				if _, err := io.ReadFull(clientConn, reply); err != nil {
					t.Fatal(err)
				}
				return reply
			}
			if reply := exchange([]byte{5, 1, 2}, 2); reply[1] != 2 {
				t.Fatalf("method selection = %v, want username/password", reply)
			}
			auth := append([]byte{1, byte(len(tc.username))}, tc.username...)
			auth = append(auth, byte(len(tc.password)))
			auth = append(auth, tc.password...)
			if reply := exchange(auth, 2); (reply[1] != 0) != tc.denied {
				t.Fatalf("auth status = %v, want denied = %v", reply, tc.denied)
			}
			if tc.denied {
				return
			}
			exchange([]byte{5, 1, 0, 1, 127, 0, 0, 1, 0, 80}, 10)
		})
	}
}
