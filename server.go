package gosocks5

import (
	"log"
	"net"
)

type Server struct {
	Addr string // TCP address to listen on

	SelectMethod   func(methods ...uint8) uint8
	MethodSelected func(method uint8, conn net.Conn) (net.Conn, error)
	Handle         func(conn net.Conn)
}

func (s *Server) ListenAndServe() error {
	addr, err := net.ResolveTCPAddr("tcp", s.Addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		//log.Println("accept", conn.RemoteAddr())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	methods, err := ReadMethods(conn)
	if err != nil {
		log.Println(err)
		return
	}

	method := MethodNoAuth
	if s.SelectMethod != nil {
		method = s.SelectMethod(methods...)
	}

	if _, err := conn.Write([]byte{Ver5, method}); err != nil {
		log.Println(err)
		return
	}

	if s.MethodSelected != nil {
		c, err := s.MethodSelected(method, conn)
		if err != nil {
			log.Println(err)
			return
		}
		conn = c
	}

	if s.Handle != nil {
		s.Handle(conn)
	}
}
