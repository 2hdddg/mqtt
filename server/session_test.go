package server

import (
	"testing"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

func TestReceivePingRequest(t *testing.T) {
	sess, conn, _ := tSession(t)
	conn.tWritePacket(&packet.PingReq{})
	// Wait for ping response or hang
	x := <-conn.written
	_ = x.(*packet.PingResp)
	sess.Stop()
}

func TestReceivePublishQoS0(t *testing.T) {
	sess, conn, p := tSession(t)
	conn.tWritePacket(&packet.Publish{
		QoS: packet.QoS0,
	})
	// Wait for publish callback or hang
	<-p.publishChan
	sess.Stop()
}

func TestReceivePublishQoS1(t *testing.T) {
	sess, conn, p := tSession(t)
	conn.tWritePacket(&packet.Publish{
		QoS: packet.QoS1,
	})
	// Wait for publish ack
	x := <-conn.written
	_ = x.(*packet.PublishAck)
	// Wait for publish callback or hang
	<-p.publishChan
	sess.Stop()
}

func TestReceiveSubscribe(t *testing.T) {
	sess, conn, _ := tSession(t)
	sub := &packet.Subscribe{
		PacketId: 0x0666,
		Subscriptions: []packet.Subscription{
			packet.Subscription{
				Topic: "a/b",
				QoS:   packet.QoS0,
			},
		},
	}
	conn.tWritePacket(sub)
	// Wait for subscribe ack
	x := <-conn.written
	_ = x.(*packet.SubscribeAck)
	sess.Stop()
}

func TestReceiveDisconnect(t *testing.T) {
	sess, conn, _ := tSession(t)
	conn.tWritePacket(&packet.Disconnect{})
	// TODO: Assert what happens on disconnect!
	sess.Stop()
}

func TestEvalPublish(t *testing.T) {
	// Subscribe first, otherwise nothing will be received.
	sess, conn, _ := tSession(t)
	sub := &packet.Subscribe{
		PacketId: 0x0666,
		Subscriptions: []packet.Subscription{
			packet.Subscription{
				Topic: "a/#",
				QoS:   packet.QoS0,
			},
		},
	}
	conn.tWritePacket(sub)
	<-conn.written // Wait for subscribe ack

	topic := topic.NewName("a/b")
	publish := &packet.Publish{}
	sess.EvalPublish(topic, publish)
	// Since topic matches subscription we should get the published.
	x := <-conn.written
	_ = x.(*packet.Publish)
	sess.Stop()
}
