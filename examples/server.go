package main

import (
	"flag"
	"log"
	"net"

	"github.com/ginuerzh/gosocks5/server"
)

var (
	laddr string
)

func init() {
	flag.StringVar(&laddr, "l", ":1080", "SOCKS5 server address")
	flag.Parse()
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := &server.Server{
		Listener: ln,
	}

	log.Fatal(srv.Serve(server.DefaultHandler))
}
