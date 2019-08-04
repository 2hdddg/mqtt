package server

import (
	"testing"

	"github.com/2hdddg/mqtt/packet"
)

func TestPingRequest(t *testing.T) {
	sess, r, w, _ := tConnect(t)
	r.tWritePacket(&packet.PingReq{})
	// Wait for ping response or hang
	<-w.pingRespChan
	sess.Stop()
}

func TestClientPublishQoS0(t *testing.T) {
	sess, r, _, p := tConnect(t)
	r.tWritePacket(&packet.Publish{})
	// Wait for publish callback or hang
	<-p.publishChan
	sess.Stop()
}

