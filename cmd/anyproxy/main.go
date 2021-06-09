package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

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

	matches := []string{
		"http://" + address,
		"socks4://" + address,
		"socks5://" + address,
	}
	svc, err := anyproxy.NewAnyProxy(context.Background(), matches, &dialer, logger, nil)
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Println("listen %s", matches)
	err = svc.ListenAndServe("tcp", address)
	if err != nil {
		logger.Println(err)
	}
}
