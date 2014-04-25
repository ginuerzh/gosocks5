// SOCKS Protocol Version 5
// http://tools.ietf.org/html/rfc1928
package gosocks5

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const Version5 uint8 = 5

type MethodType uint8

const (
	MethodNoAuth MethodType = iota
	MethodGSSAPI
	MethodUserPass
	// X'03' to X'7F' IANA ASSIGNED
	// X'80' to X'FE' RESERVED FOR PRIVATE METHODS
	MethodNoAcceptable = 0xFF
)

type CmdType uint8

const (
	CmdConnect CmdType = 1
	CmdBind            = 2
	CmdUdp             = 3
)

type AddrType uint8

const (
	AddrIPv4       AddrType = 1
	AddrDomainName          = 3
	AddrIPv6                = 4
)

type ReplyType uint8

const (
	Succeeded ReplyType = iota
	Failure
	NotAllowed
	NetUnreachable
	HostUnreachable
	ConnRefused
	TTLExpired
	CmdUnsupported
	AddrUnsupported
)

type Socks5 struct {
	conn    net.Conn
	methods Methods
}

func NewSocks5(conn net.Conn, methods ...MethodType) *Socks5 {
	s := &Socks5{
		conn:    conn,
		methods: Methods(methods),
	}
	if len(s.methods) == 0 {
		s.methods = append(s.methods, MethodNoAuth)
	}
	return s
}

func (s *Socks5) Init() error {
	b := make([]byte, 2)
	if _, err := s.conn.Write(s.methods.Encode()); err != nil {
		return err
	}
	if _, err := s.conn.Read(b); err != nil {
		return err
	}
	return nil
}

/*
+----+----------+----------+
|VER | NMETHODS | METHODS  |
+----+----------+----------+
| 1  |    1     | 1 to 255 |
+----+----------+----------+
*/
type Methods []MethodType

func (methods Methods) Encode() []byte {
	b := make([]byte, 257)
	pos := 0
	b[pos] = Version5
	pos++
	b[pos] = byte(len(methods))
	pos++
	for _, m := range methods {
		b[pos] = byte(m)
		pos++
	}

	return b[:pos]
}

/*
+----+-----+-------+------+----------+----------+
|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
+----+-----+-------+------+----------+----------+
| 1  |  1  | X'00' |  1   | Variable |    2     |
+----+-----+-------+------+----------+----------+
*/
type CMD struct {
	Cmd   CmdType
	AType AddrType
	Addr  string
	Port  uint16
}

func NewCMD(cmdType CmdType, aType AddrType, addr string, port uint16) *CMD {
	return &CMD{
		Cmd:   cmdType,
		AType: aType,
		Addr:  addr,
		Port:  port,
	}
}

func (cmd *CMD) Encode() []byte {
	b := make([]byte, 128)
	b[0] = Version5
	b[1] = byte(cmd.Cmd)
	b[3] = byte(cmd.AType)
	pos := 4

	switch cmd.AType {
	case AddrIPv4:
		pos += copy(b[pos:], net.ParseIP(cmd.Addr).To4())
	case AddrDomainName:
		b[pos] = byte(len(cmd.Addr))
		pos++
		pos += copy(b[pos:], []byte(cmd.Addr))
	case AddrIPv6:
		pos += copy(b[pos:], net.ParseIP(cmd.Addr).To16())
	}
	binary.BigEndian.PutUint16(b[pos:], cmd.Port)

	return b[:pos+2]
}

func (cmd *CMD) Decode(data []byte) {
	cmd.Cmd = CmdType(data[1])
	cmd.AType = AddrType(data[3])

	pos := 4
	switch cmd.AType {
	case AddrIPv4:
		cmd.Addr = net.IP(data[pos : pos+4]).String()
		pos += 4
	case AddrDomainName:
		length := int(data[pos])
		pos++
		cmd.Addr = string(data[pos : pos+length])
		pos += length
	case AddrIPv6:
		cmd.Addr = net.IP(data[pos : pos+16]).String()
		pos += 16
	}

	cmd.Port = binary.BigEndian.Uint16(data[pos:])
}

func (cmd *CMD) String() string {
	b := bytes.Buffer{}
	b.WriteString("cmd:" + fmt.Sprintf("%x", cmd.Cmd))
	b.WriteString("\natype:" + fmt.Sprintf("%x", cmd.AType))
	b.WriteString("\naddr:" + cmd.Addr)
	b.WriteString("\nport:" + fmt.Sprintf("%x", cmd.Port))
	return b.String()
}
