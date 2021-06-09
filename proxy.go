package anyproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"sort"

	"github.com/wzshiming/cmux"
	"github.com/wzshiming/cmux/pattern"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type AnyProxy struct {
	proxies map[string]*Host
}

func NewAnyProxy(ctx context.Context, addrs []string, dial Dialer, logger *log.Logger, pool BytesPool) (*AnyProxy, error) {
	proxies := map[string]*Host{}
	for _, addr := range addrs {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}
		host := u.Host

		s, err := newServer(ctx, addr, dial, logger, pool)
		if err != nil {
			return nil, err
		}
		mux, ok := proxies[host]
		if !ok {
			mux = &Host{
				cmux: cmux.NewCMux(),
			}
		}
		patterns := s.Patterns()
		if patterns == nil {
			mux.proxies = append(mux.proxies, s.ProxyURL())
			err = mux.cmux.NotFound(s)
			if err != nil {
				return nil, err
			}
		} else {
			mux.proxies = append(mux.proxies, s.ProxyURL())
			for _, p := range patterns {
				err = mux.cmux.HandlePrefix(s, pattern.Pattern[p]...)
				if err != nil {
					return nil, err
				}
			}
		}
		proxies[u.Host] = mux
	}
	proxy := &AnyProxy{
		proxies: proxies,
	}
	return proxy, nil
}

func (a *AnyProxy) Match(addr string) *Host {
	return a.proxies[addr]
}

func (a *AnyProxy) Hosts() []string {
	hosts := make([]string, 0, len(a.proxies))
	for proxy := range a.proxies {
		hosts = append(hosts, proxy)
	}
	sort.Strings(hosts)
	return hosts
}

func (a *AnyProxy) ListenAndServe(network, address string) error {
	host := a.Match(address)
	if host == nil {
		return fmt.Errorf("not match address %q", address)
	}
	return host.ListenAndServe(network, address)
}

type Host struct {
	cmux    *cmux.CMux
	proxies []string
}

func (h *Host) ProxyURLs() []string {
	return h.proxies
}

func (h *Host) ServeConn(conn net.Conn) {
	h.cmux.ServeConn(conn)
}

func (h *Host) ListenAndServe(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go h.ServeConn(conn)
	}
}
