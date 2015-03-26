package gosocks5

import (
	"io"
	"net"
)

type Client struct {
	Methods []uint8
	Conn    net.Conn
}

func (c *Client) Handshake() (method uint8, err error) {
	nm := len(c.Methods)
	if nm == 0 {
		nm = 1
	}
	b := make([]byte, 2+nm)

	if _, err = c.Conn.Write([]byte{Ver5, 1, 0}); err != nil {
		return
	}

	if _, err = io.ReadFull(c.Conn, b[:2]); err != nil {
		return
	}

	method = b[1]

	if b[0] != Ver5 {
		err = ErrBadVersion
		return
	}

	return
}

func (c *Client) Request(r *Request) (*Reply, error) {
	if err := r.Write(c.Conn); err != nil {
		return nil, err
	}

	return ReadReply(c.Conn)
}
