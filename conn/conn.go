package conn

import (
	"bufio"
	"net"

	"github.com/2hdddg/mqtt/packet"
)

type Conn struct {
	net.Conn
	*packet.Reader
	*packet.Writer
}

func New(c net.Conn) *Conn {
	rd := &packet.Reader{bufio.NewReader(c)}
	wr := &packet.Writer{c}
	return &Conn{c, rd, wr}
}
