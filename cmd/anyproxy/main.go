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
	svc := anyproxy.NewAnyProxy(context.Background(), &dialer, logger, nil)

	err := svc.ListenAndServe("tcp", address)
	if err != nil {
		logger.Println(err)
	}
}
