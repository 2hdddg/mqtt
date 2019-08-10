package server

import (
	"testing"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

func TestReceivePingRequest(t *testing.T) {
	sess, r, w, _ := tSession(t)
	r.tWritePacket(&packet.PingReq{})
	// Wait for ping response or hang
	x := <-w.written
	_ = x.(*packet.PingResp)
	sess.Stop()
}

func TestReceivePublishQoS0(t *testing.T) {
	sess, r, _, p := tSession(t)
	r.tWritePacket(&packet.Publish{})
	// Wait for publish callback or hang
	<-p.publishChan
	sess.Stop()
}

func TestReceiveSubscribe(t *testing.T) {
	sess, rd, wr, _ := tSession(t)
	sub := &packet.Subscribe{
		PacketId: 0x0666,
		Subscriptions: []packet.Subscription{
			packet.Subscription{
				Topic: "a/b",
				QoS:   packet.QoS0,
			},
		},
	}
	rd.tWritePacket(sub)
	// Wait for subscribe ack
	x := <-wr.written
	_ = x.(*packet.SubscribeAck)
	sess.Stop()
}

func TestReceiveDisconnect(t *testing.T) {
	sess, rd, _, _ := tSession(t)
	rd.tWritePacket(&packet.Disconnect{})
	// TODO: Assert what happens on disconnect!
	sess.Stop()
}

func TestEvalPublish(t *testing.T) {
	// Subscribe first, otherwise nothing will be received.
	sess, rd, wr, _ := tSession(t)
	sub := &packet.Subscribe{
		PacketId: 0x0666,
		Subscriptions: []packet.Subscription{
			packet.Subscription{
				Topic: "a/#",
				QoS:   packet.QoS0,
			},
		},
	}
	rd.tWritePacket(sub)
	<-wr.written // Wait for subscribe ack

	topic := topic.NewName("a/b")
	publish := &packet.Publish{}
	sess.EvalPublish(topic, publish)
	// Since topic matches subscription we should get the published.
	x := <-wr.written
	_ = x.(*packet.Publish)
	sess.Stop()
}
