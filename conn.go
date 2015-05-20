package gosocks5

import (
	"io"
	//"log"
	"net"
	"sync"
	"time"
)

type Config struct {
	Methods        []uint8
	SelectMethod   func(methods ...uint8) uint8
	MethodSelected func(method uint8, conn net.Conn) (net.Conn, error)
}

func defaultConfig() *Config {
	return &Config{}
}

type Conn struct {
	c              net.Conn
	config         *Config
	method         uint8
	isClient       bool
	handshaked     bool
	handshakeMutex sync.Mutex
	handshakeErr   error
}

func ClientConn(conn net.Conn, config *Config) *Conn {
	return &Conn{
		c:        conn,
		config:   config,
		isClient: true,
	}
}

func ServerConn(conn net.Conn, config *Config) *Conn {
	return &Conn{
		c:      conn,
		config: config,
	}
}

func (conn *Conn) Handleshake() error {
	conn.handshakeMutex.Lock()
	defer conn.handshakeMutex.Unlock()

	if err := conn.handshakeErr; err != nil {
		return err
	}
	if conn.handshaked {
		return nil
	}

	if conn.isClient {
		conn.handshakeErr = conn.clientHandshake()
	} else {
		conn.handshakeErr = conn.serverHandshake()
	}

	return conn.handshakeErr
}

func (conn *Conn) clientHandshake() error {
	if conn.config == nil {
		conn.config = defaultConfig()
	}

	nm := len(conn.config.Methods)
	if nm == 0 {
		nm = 1
	}

	b := make([]byte, 2+nm)
	b[0] = Ver5
	b[1] = uint8(nm)
	copy(b[2:], conn.config.Methods)

	if _, err := conn.c.Write(b); err != nil {
		return err
	}

	if _, err := io.ReadFull(conn.c, b[:2]); err != nil {
		return err
	}

	if b[0] != Ver5 {
		return ErrBadVersion
	}

	if conn.config.MethodSelected != nil {
		c, err := conn.config.MethodSelected(b[1], conn.c)
		if err != nil {
			return err
		}
		conn.c = c
	}
	conn.method = b[1]
	//log.Println("method:", conn.method)
	conn.handshaked = true
	return nil
}

func (conn *Conn) serverHandshake() error {
	if conn.config == nil {
		conn.config = defaultConfig()
	}

	methods, err := ReadMethods(conn.c)
	if err != nil {
		return err
	}

	method := MethodNoAuth
	if conn.config.SelectMethod != nil {
		method = conn.config.SelectMethod(methods...)
	}

	if _, err := conn.c.Write([]byte{Ver5, method}); err != nil {
		return err
	}

	if conn.config.MethodSelected != nil {
		c, err := conn.config.MethodSelected(method, conn.c)
		if err != nil {
			return err
		}
		conn.c = c
	}
	conn.method = method
	//log.Println("method:", method)
	conn.handshaked = true
	return nil
}

func (conn *Conn) Read(b []byte) (n int, err error) {
	if err = conn.Handleshake(); err != nil {
		return
	}
	return conn.c.Read(b)
}

func (conn *Conn) Write(b []byte) (n int, err error) {
	if err = conn.Handleshake(); err != nil {
		return
	}
	return conn.c.Write(b)
}

func (conn *Conn) Close() error {
	return conn.c.Close()
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.c.LocalAddr()
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.c.RemoteAddr()
}

func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.c.SetDeadline(t)
}

func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.c.SetReadDeadline(t)
}

func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.c.SetWriteDeadline(t)
}
