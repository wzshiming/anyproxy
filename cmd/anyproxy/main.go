package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	_ "github.com/wzshiming/anyproxy/init"
	_ "github.com/wzshiming/anyproxy/pprof"

	"github.com/wzshiming/anyproxy"
)

var address string

func init() {
	flag.StringVar(&address, "a", ":8080", "listen on the address")
	flag.Parse()
}

func main() {
	logger := log.New(os.Stderr, "[any proxy] ", log.LstdFlags)
	var dialer net.Dialer

	addrs := flag.Args()
	if address != "" {
		addrs = append(addrs, "http://"+address, "socks4://"+address, "socks5://"+address, "ssh://"+address, "pprof://"+address)
	}

	conf := anyproxy.Config{
		Dialer: &dialer,
		Logger: logger,
	}

	svc, err := anyproxy.NewAnyProxy(context.Background(), addrs, &conf)
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Printf("listen %s", addrs)
	err = svc.Run(context.Background())
	if err != nil {
		logger.Println(err)
	}
}
