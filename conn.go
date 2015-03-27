package gosocks5

import (
	//"io"
	"net"
)

type Conn struct {
	net.Conn
}

func Dial(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{Conn: conn}, nil
}
