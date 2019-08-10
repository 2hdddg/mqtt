package server

import (
	"net"
	"testing"
	"time"

	"github.com/2hdddg/mqtt/packet"
)

type ReaderFake struct {
	pack chan interface{}
	err  chan error
}

func tNewReaderFake(t *testing.T) *ReaderFake {
	return &ReaderFake{
		pack: make(chan interface{}, 1),
		err:  make(chan error, 1),
	}
}

// Implements Reader interface
func (r *ReaderFake) ReadPacket(version uint8) (interface{}, error) {
	for {
		select {
		case p := <-r.pack:
			return p, nil
		case e := <-r.err:
			return nil, e
		}
	}
}

func (r *ReaderFake) tWritePacket(p interface{}) {
	r.pack <- p
}

type WriterFake struct {
	err     error
	written chan interface{}
}

func (w *WriterFake) WritePacket(p interface{}) error {
	w.written <- p
	return w.err
}

type ConnFake struct {
	closed bool
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
func (c *ConnFake) LocalAddr() net.Addr {
	return nil
}
func (c *ConnFake) RemoteAddr() net.Addr {
	return nil
}
func (c *ConnFake) SetDeadline(t time.Time) error {
	return nil
}
func (c *ConnFake) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *ConnFake) SetWriteDeadline(t time.Time) error {
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

func tSession(
	t *testing.T) (*Session, *ReaderFake, *WriterFake, *PubFake) {

	rd := tNewReaderFake(t)
	connect := &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		KeepAliveSecs:    30,
		ClientIdentifier: "xyz",
	}
	wr := &WriterFake{
		written: make(chan interface{}, 3),
	}
	conn := &ConnFake{}
	pub := NewPubFake()
	sess := newSession(conn, rd, wr, connect)
	sess.Start(pub)
	return sess, rd, wr, pub
}

