package server

import (
	"errors"
	"net"
	"testing"
	"time"

	packet "github.com/2hdddg/mqtt/controlpacket"
)

type ReaderFake struct {
	pack interface{}
	err  error
}

func (r *ReaderFake) ReadPacket(version uint8) (interface{}, error) {
	p := r.pack
	r.pack = nil
	return p, r.err
}

type WriterFake struct {
	ack          *packet.AckConnection
	acks         int
	ackErr       error
	pingResps    int
	pingRespChan chan bool
}

func (w *WriterFake) WriteAckConnection(ack *packet.AckConnection) error {
	w.ack = ack
	w.acks++
	return w.ackErr
}

func (w *WriterFake) WritePingResp(resp *packet.PingResp) error {
	w.pingResps++
	if w.pingRespChan != nil {
		w.pingRespChan <- true
	}
	return nil
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

func TestConnect(t *testing.T) {
	testcases := []struct {
		connectPacket *packet.Connect
		readError     error
		shouldFail    bool
		shouldClose   bool
		shouldAck     bool
		ackRetCode    packet.ConnRetCode
		writeAckErr   error
	}{
		// Happy path, ok protocol and version
		{
			&packet.Connect{
				ProtocolName:     "MQTT",
				ProtocolVersion:  4,
				KeepAliveSecs:    30,
				ClientIdentifier: "xyz",
			}, nil, false, false, true, packet.ConnAccepted, nil,
		},
		// Fail to send ack
		{
			&packet.Connect{
				ProtocolName:     "MQTT",
				ProtocolVersion:  4,
				KeepAliveSecs:    30,
				ClientIdentifier: "xyz",
			}, nil, true, true, true, packet.ConnAccepted,
			errors.New(""),
		},
		// Unknown protocol
		{
			&packet.Connect{
				ProtocolName:     "Unknown",
				ProtocolVersion:  4,
				KeepAliveSecs:    30,
				ClientIdentifier: "xyz",
			}, nil, true, true, false, packet.ConnRefusedVersion, nil,
		},
		// Wrong version
		{
			&packet.Connect{
				ProtocolName:     "MQTT",
				ProtocolVersion:  1,
				KeepAliveSecs:    30,
				ClientIdentifier: "xyz",
			}, nil, true, true, true, packet.ConnRefusedVersion, nil,
		},
	}

	for _, c := range testcases {
		r := &ReaderFake{}
		w := &WriterFake{
			ackErr: c.writeAckErr,
		}
		conn := &ConnFake{}
		au := &AuthFake{}

		r.pack = c.connectPacket
		r.err = c.readError
		sess, err := Connect(conn, r, w, au)
		if c.shouldFail && (sess != nil || err == nil) {
			t.Errorf("Should fail")
		}
		if !c.shouldFail && (sess == nil || err != nil) {
			t.Errorf("Should succeed")
		}
		if c.shouldClose && !conn.closed {
			t.Errorf("Should have closed")
		}
		if !c.shouldClose && conn.closed {
			t.Errorf("Should NOT have closed")
		}
		if c.shouldAck && w.acks != 1 {
			t.Errorf("Should have acked")
		}
		if c.shouldAck && w.ack.RetCode != c.ackRetCode {
			t.Errorf("Wrong ack")
		}
	}
}
