package client

import (
	"net"
	"time"

	"github.com/ginuerzh/gosocks5"
)

type Client struct {
	selector gosocks5.Selector
}

func (c *Client) Dial(addr string, options ...DialOption) (net.Conn, error) {
	return nil, nil
}

// DialOptions describes the options for Transporter.Dial.
type DialOptions struct {
	Timeout time.Duration
}

// DialOption allows a common way to set dial options.
type DialOption func(opts *DialOptions)
