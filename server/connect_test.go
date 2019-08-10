package server

import (
	"errors"
	"testing"

	"github.com/2hdddg/mqtt/packet"
)

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
		r := tNewReaderFake(t)
		w := &WriterFake{
			err:     c.writeAckErr,
			written: make(chan packet.Packet, 3),
		}
		conn := &ConnFake{}
		au := &AuthFake{}

		if c.connectPacket != nil {
			r.pack <- c.connectPacket
		}
		if c.readError != nil {
			r.err <- c.readError
		}
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
		if c.shouldAck {
			x := <-w.written
			ack := x.(*packet.AckConnection)
			if ack.RetCode != c.ackRetCode {
				t.Errorf("Wrong ack")
			}
		}
	}
}
