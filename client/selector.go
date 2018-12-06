package client

import (
	"net"
	"net/url"

	"github.com/ginuerzh/gosocks5"
)

var (
	// DefaultSelector is the default client selector.
	DefaultSelector gosocks5.Selector = &clientSelector{}
)

type clientSelector struct {
	methods []uint8
	user    *url.Userinfo
}

func NewClientSelector(user *url.Userinfo, methods ...uint8) gosocks5.Selector {
	return &clientSelector{
		methods: methods,
		user:    user,
	}
}

func (selector *clientSelector) Methods() []uint8 {
	return selector.methods
}

func (selector *clientSelector) AddMethod(methods ...uint8) {
	selector.methods = append(selector.methods, methods...)
}

func (selector *clientSelector) Select(methods ...uint8) (method uint8) {
	return
}

func (selector *clientSelector) OnSelected(method uint8, conn net.Conn) (net.Conn, error) {
	switch method {
	case gosocks5.MethodUserPass:
		var username, password string
		if selector.user != nil {
			username = selector.user.Username()
			password, _ = selector.user.Password()
		}

		req := gosocks5.NewUserPassRequest(gosocks5.UserPassVer, username, password)
		if err := req.Write(conn); err != nil {
			return nil, err
		}
		resp, err := gosocks5.ReadUserPassResponse(conn)
		if err != nil {
			return nil, err
		}
		if resp.Status != gosocks5.Succeeded {
			return nil, gosocks5.ErrAuthFailure
		}
	case gosocks5.MethodNoAcceptable:
		return nil, gosocks5.ErrBadMethod
	}

	return conn, nil
}
