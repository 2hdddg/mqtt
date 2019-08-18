package conn

import (
	"bufio"
	"net"
	"time"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

type Conn struct {
	c net.Conn
	*packet.Reader
	*packet.Writer
	closed bool
	log    logger.L
}

func New(c net.Conn) *Conn {
	rd := &packet.Reader{bufio.NewReader(c)}
	wr := &packet.Writer{c}
	return &Conn{c, rd, wr, false, nil}
}

func (c *Conn) Close() error {
	if c.log != nil {
		c.log.Debug("Connection closed")
	}
	c.closed = true
	return c.c.Close()
}

func (c *Conn) SetLog(log logger.L) {
	c.log = log
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *Conn) IsClosed() bool {
	return c.closed
}
