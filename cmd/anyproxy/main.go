package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	_ "github.com/wzshiming/anyproxy/init"

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

	args := flag.Args()
	if address != "" {
		args = append(args, "http://"+address, "socks4://"+address, "socks5://"+address)
	}
	svc, err := anyproxy.NewAnyProxy(context.Background(), args, &dialer, logger, nil)
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Printf("listen %s", args)
	err = svc.Run(context.Background())
	if err != nil {
		logger.Println(err)
	}
}
