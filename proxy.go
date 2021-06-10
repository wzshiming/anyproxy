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
	"golang.org/x/sync/errgroup"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type AnyProxy struct {
	proxies map[string]*Host
}

func NewAnyProxy(ctx context.Context, addrs []string, dial Dialer, logger *log.Logger, pool BytesPool) (*AnyProxy, error) {
	proxies := map[string]*Host{}
	users := map[string][]*url.Userinfo{}
	for _, addr := range addrs {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}

		unique := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		if u.User == nil {
			users[unique] = nil
		} else {
			if user, ok := users[unique]; !ok || user != nil {
				users[unique] = append(users[unique], u.User)
			}
		}
		sch, ok := schemeMap[u.Scheme]
		if !ok {
			return nil, fmt.Errorf("can't support scheme %q", u.Scheme)
		}

		s, patterns, err := newServer(ctx, sch, u.Host, users[unique], dial, logger, pool)
		if err != nil {
			return nil, err
		}
		mux, ok := proxies[u.Host]
		if !ok {
			mux = &Host{
				cmux: cmux.NewCMux(),
			}
		}
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

func (a *AnyProxy) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	for _, address := range a.Hosts() {
		address := address
		host := a.Match(address)
		if host == nil {
			return fmt.Errorf("not match address %q", address)
		}
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		group.Go(func() error {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return err
				}
				go host.ServeConn(conn)
			}
		})
	}
	return group.Wait()
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
