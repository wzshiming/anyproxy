package pprof

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"

	"github.com/wzshiming/anyproxy"
	"github.com/wzshiming/cmux/pattern"
)

const prefix = "/debug/pprof/"

func NewServeConn(ctx context.Context, sch, address string, users []*url.Userinfo, dial anyproxy.Dialer, logger anyproxy.Logger, pool anyproxy.BytesPool) (anyproxy.ServeConn, []string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc(prefix+"", pprof.Index)
	mux.HandleFunc(prefix+"cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"profile", pprof.Profile)
	mux.HandleFunc(prefix+"symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"trace", pprof.Trace)

	var patterns []string

	tmp := pattern.Pattern[pattern.HTTP]
	patterns = make([]string, 0, len(tmp)+1)
	for _, t := range tmp {
		patterns = append(patterns, t+"/debug/")
	}
	s := http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		Handler: mux,
	}
	return anyproxy.NewHttpServeConn(&s), patterns, nil
}
