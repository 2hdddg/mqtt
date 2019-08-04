package server

import (
	"fmt"
	"testing"

	"github.com/2hdddg/mqtt/packet"
)

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

func tConnect(t *testing.T) (*Session, *ReaderFake, *WriterFake, *PubFake) {
	fmt.Println("tConnect")
	r := NewReaderFake()
	r.pack <- &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		KeepAliveSecs:    30,
		ClientIdentifier: "xyz",
	}
	w := &WriterFake{
		pingRespChan: make(chan bool),
	}
	au := &AuthFake{}
	conn := &ConnFake{}
	pub := NewPubFake()
	// Connect to get a proper session
	sess, _ := Connect(conn, r, w, au)
	if sess == nil {
		t.Fatalf("Could not connect to create session")
	}
	sess.Start(pub)
	return sess, r, w, pub
}

func TestPingRequest(t *testing.T) {
	sess, r, w, _ := tConnect(t)
	r.pack <- &packet.PingReq{}
	// Wait for ping response or hang
	<-w.pingRespChan
	sess.Stop()
}

func TestClientPublishQoS0(t *testing.T) {
	sess, r, _, p := tConnect(t)
	r.pack <- &packet.Publish{}
	// Wait for publish callback or hang
	<-p.publishChan
	sess.Stop()
}

