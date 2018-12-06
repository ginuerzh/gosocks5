package main

import (
	"flag"
	"log"

	"github.com/ginuerzh/gosocks5"

	"github.com/ginuerzh/gosocks5/client"
)

var (
	server string
)

func init() {
	flag.StringVar(&server, "p", "", "SOCKS5 server address")
	flag.Parse()
}

func main() {
	addr, err := gosocks5.NewAddr(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := client.Dial(server)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := gosocks5.NewRequest(gosocks5.CmdConnect, addr).Write(conn); err != nil {
		log.Fatal(err)
	}
	reply, err := gosocks5.ReadReply(conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reply:", reply)
}
