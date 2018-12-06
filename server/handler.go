package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/ginuerzh/gosocks5"
)

var (
	// DefaultHandler is the default server handler.
	DefaultHandler Handler
)

func init() {
	DefaultHandler = &serverHandler{
		selector: DefaultSelector,
	}
}

// Handler is interface for server handler.
type Handler interface {
	Handle(conn net.Conn) error
}

type serverHandler struct {
	selector gosocks5.Selector
}

func (h *serverHandler) Handle(conn net.Conn) error {
	conn = gosocks5.ServerConn(conn, h.selector)
	req, err := gosocks5.ReadRequest(conn)
	if err != nil {
		return err
	}

	switch req.Cmd {
	case gosocks5.CmdConnect:
		return h.handleConnect(conn, req)

	case gosocks5.CmdBind:
		return h.handleBind(conn, req)

	// case gosocks5.CmdUdp:
	// h.handleUDPRelay(conn, req)

	default:
		return fmt.Errorf("%d: unsupported command", gosocks5.CmdUnsupported)
	}
}

func (h *serverHandler) handleConnect(conn net.Conn, req *gosocks5.Request) error {
	cc, err := net.Dial("tcp", req.Addr.String())
	if err != nil {
		rep := gosocks5.NewReply(gosocks5.HostUnreachable, nil)
		rep.Write(conn)
		return err
	}
	defer cc.Close()

	rep := gosocks5.NewReply(gosocks5.Succeeded, nil)
	if err := rep.Write(conn); err != nil {
		return err
	}

	return transport(conn, cc)
}

var (
	trPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1500)
		},
	}
)

func (h *serverHandler) handleBind(conn net.Conn, req *gosocks5.Request) error {
	addr := req.Addr.String()
	bindAddr, _ := net.ResolveTCPAddr("tcp", addr)
	ln, err := net.ListenTCP("tcp", bindAddr) // strict mode: if the port already in use, it will return error
	if err != nil {
		gosocks5.NewReply(gosocks5.Failure, nil).Write(conn)
		return err
	}

	socksAddr := toSocksAddr(ln.Addr())
	// Issue: may not reachable when host has multi-interface
	socksAddr.Host, _, _ = net.SplitHostPort(conn.LocalAddr().String())
	reply := gosocks5.NewReply(gosocks5.Succeeded, socksAddr)
	if err := reply.Write(conn); err != nil {
		ln.Close()
		return err
	}

	var pconn net.Conn
	accept := func() <-chan error {
		errc := make(chan error, 1)

		go func() {
			defer close(errc)
			defer ln.Close()

			c, err := ln.AcceptTCP()
			if err != nil {
				errc <- err
				return
			}
			pconn = c
		}()

		return errc
	}

	pc1, pc2 := net.Pipe()
	pipe := func() <-chan error {
		errc := make(chan error, 1)

		go func() {
			defer close(errc)
			defer pc1.Close()

			errc <- transport(conn, pc1)
		}()

		return errc
	}

	defer pc2.Close()

	for {
		select {
		case err := <-accept():
			if err != nil || pconn == nil {
				return err
			}
			defer pconn.Close()

			reply := gosocks5.NewReply(gosocks5.Succeeded, toSocksAddr(pconn.RemoteAddr()))
			if err := reply.Write(pc2); err != nil {
				return err
			}

			if err = transport(pc2, pconn); err != nil {
			}

			return err
		case err := <-pipe():
			ln.Close()
			return err
		}
	}
}

func transport(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		buf := trPool.Get().([]byte)
		defer trPool.Put(buf)

		_, err := io.CopyBuffer(rw1, rw2, buf)
		errc <- err
	}()

	go func() {
		buf := trPool.Get().([]byte)
		defer trPool.Put(buf)

		_, err := io.CopyBuffer(rw2, rw1, buf)
		errc <- err
	}()

	err := <-errc
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func toSocksAddr(addr net.Addr) *gosocks5.Addr {
	host := "0.0.0.0"
	port := 0
	if addr != nil {
		h, p, _ := net.SplitHostPort(addr.String())
		host = h
		port, _ = strconv.Atoi(p)
	}
	return &gosocks5.Addr{
		Type: gosocks5.AddrIPv4,
		Host: host,
		Port: uint16(port),
	}
}
