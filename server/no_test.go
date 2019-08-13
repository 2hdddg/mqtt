package server

import (
	"testing"
	"time"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

type ConnFake struct {
	rdpack  chan packet.Packet
	rderr   chan error
	wrerr   error
	written chan packet.Packet
	closed  bool
}

func tNewConnFake(t *testing.T) *ConnFake {
	return &ConnFake{
		rdpack:  make(chan packet.Packet, 1),
		rderr:   make(chan error, 1),
		written: make(chan packet.Packet, 3),
	}
}

func (r *ConnFake) ReadPacket(version uint8, log logger.L) (packet.Packet, error) {
	for {
		select {
		case p := <-r.rdpack:
			return p, nil
		case e := <-r.rderr:
			return nil, e
		}
	}
}

func (r *ConnFake) tWritePacket(p packet.Packet) {
	r.rdpack <- p
}

func (w *ConnFake) WritePacket(p packet.Packet, log logger.L) error {
	w.written <- p
	return w.wrerr
}

func (c *ConnFake) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (c *ConnFake) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (c *ConnFake) Close() error {
	c.closed = true
	return nil
}
func (c *ConnFake) SetReadDeadline(t time.Time) error {
	return nil
}

type AuthFake struct {
}

func (a *AuthFake) CheckConnect(c *packet.Connect) packet.ConnRetCode {
	return packet.ConnAccepted
}

type PubFake struct {
	publishChan chan *packet.Publish
}

func NewPubFake() *PubFake {
	return &PubFake{
		publishChan: make(chan *packet.Publish, 1),
	}
}

func (f *PubFake) Publish(s *Session, p *packet.Publish) error {
	f.publishChan <- p
	return nil
}

func (f *PubFake) Stopped(s *Session) {
}

func tSession(
	t *testing.T) (*Session, *ConnFake, *PubFake) {

	connect := &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		KeepAliveSecs:    30,
		ClientIdentifier: "xyz",
	}
	conn := tNewConnFake(t)
	pub := NewPubFake()
	log := logger.NewSession("123")
	sess := NewSession(conn, connect, pub, log)
	return sess, conn, pub
}

