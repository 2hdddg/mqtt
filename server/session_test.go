package server

import (
	"testing"

	packet "github.com/2hdddg/mqtt/controlpacket"
)

func TestPingRequest(t *testing.T) {
	r := &ReaderFake{
		pack: &packet.Connect{
			ProtocolName:     "MQTT",
			ProtocolVersion:  4,
			KeepAliveSecs:    30,
			ClientIdentifier: "xyz",
		},
	}
	w := &WriterFake{
		pingRespChan: make(chan bool),
	}
	conn := &ConnFake{}
	// Connect to get a proper session
	sess, _ := Connect(conn, r, w)
	if sess == nil {
		t.Fatalf("Could not connect to create session")
	}
	r.pack = &packet.PingReq{}
	sess.Start()
	// Wait for ping response or hang
	<-w.pingRespChan
	sess.Stop()
}
