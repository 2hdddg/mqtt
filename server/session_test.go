package server

import (
	"testing"

	packet "github.com/2hdddg/mqtt/controlpacket"
)

type PubFake struct{}

func (f *PubFake) Publish(s *Session, p *packet.Publish) error {

	return nil
}

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
	au := &AuthFake{}
	conn := &ConnFake{}
	pub := &PubFake{}
	// Connect to get a proper session
	sess, _ := Connect(conn, r, w, au)
	if sess == nil {
		t.Fatalf("Could not connect to create session")
	}
	r.pack = &packet.PingReq{}
	sess.Start(pub)
	// Wait for ping response or hang
	<-w.pingRespChan
	sess.Stop()
}
