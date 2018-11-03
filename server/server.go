package server

import (
	"net"
	"time"
)

// Server is a SOCKS5 server.
type Server struct {
	Listener net.Listener
}

// Addr returns the address of the server
func (s *Server) Addr() net.Addr {
	return s.Listener.Addr()
}

// Serve serves incoming requests.
func (s *Server) Serve(h Handler, options ...ServerOption) error {
	if s.Listener == nil {
		ln, err := net.ListenTCP("tcp", nil)
		if err != nil {
			return err
		}
		s.Listener = ln
	}
	if h == nil {
		h = DefaultHandler
	}

	l := s.Listener
	var tempDelay time.Duration
	for {
		conn, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		go h.Handle(conn)
	}
}

// Close closes the socks5 server
func (s *Server) Close() error {
	return s.Listener.Close()
}

// ServerOptions is options for server.
type ServerOptions struct {
}

// ServerOption allows a common way to set server options.
type ServerOption func(opts *ServerOptions)
