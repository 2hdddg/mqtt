package qos

import (
	"errors"
	"sync"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/writequeue"
)

type Accept func(p *packet.Publish)

type QoS struct {
	received     map[uint16]*packet.Publish
	sent         map[uint16]*packet.Publish
	lastPacketId uint16
	wrQueue      *writequeue.Queue
	mut          *sync.Mutex
	log          logger.L
}

func New(wrQueue *writequeue.Queue, log logger.L) *QoS {

	return &QoS{
		wrQueue:  wrQueue,
		received: make(map[uint16]*packet.Publish),
		sent:     make(map[uint16]*packet.Publish),
		mut:      &sync.Mutex{},
		log:      log,
	}
}

func (q *QoS) getUnusedPacketId() uint16 {
	q.lastPacketId += 1
	return q.lastPacketId
}

func (q *QoS) writtenPUBACK(packetId uint16) {
	q.mut.Lock()
	p, exists := q.received[packetId]
	if exists && p != nil {
		delete(q.received, packetId)
	}
	q.mut.Unlock()
}

func (q *QoS) ReceivedPublish(p *packet.Publish, acc Accept) error {
	switch p.QoS {
	case packet.QoS0:
		acc(p)
	case packet.QoS1:
		// After it has sent a PUBACK Packet the Receiver MUST treat
		// any incoming PUBLISH packet that contains the same Packet
		// Identifier as being a new publication, irrespective of the
		// setting of its DUP flag.
		q.mut.Lock()
		if _, exists := q.received[p.PacketId]; !exists {
			go acc(p)
			q.received[p.PacketId] = p
		}
		q.mut.Unlock()
		// MUST respond with a PUBACK Packet containing the Packet
		// Identifier from the incoming PUBLISH Packet, having accepted
		// ownership of the Application Message
		ack := &packet.PublishAck{PacketId: p.PacketId}
		q.wrQueue.Add(&writequeue.Item{
			Packet: ack,
			Written: func() {
				q.writtenPUBACK(p.PacketId)
			},
		})
	case packet.QoS2:
		q.log.Error(
			"Reception of PUBLISH with QoS > 1 is not implemented")
		return errors.New("QoS above 1 is not implemented")
	}
	return nil
}

func (q *QoS) ReceivedPublishAck(p *packet.PublishAck) error {
	// MUST treat the PUBLISH Packet as “unacknowledged” until it has
	// received the corresponding PUBACK packet from the receiver.
	q.mut.Lock()
	defer q.mut.Unlock()
	_, exists := q.sent[p.PacketId]
	if !exists {
		q.log.Error("Received PUBACK for unknown PUBLISH")
		return errors.New("Received PUBACK for unknown publish")
	}
	delete(q.sent, p.PacketId)
	return nil
}

func (q *QoS) SendPublish(p *packet.Publish) error {
	switch p.QoS {
	case packet.QoS0:
		// MUST send a PUBLISH packet with QoS=0, DUP=0
		p.Duplicate = false
		q.wrQueue.Add(&writequeue.Item{
			Packet: p,
		})
	case packet.QoS1:
		// MUST assign an unused Packet Identifier each time it has a
		// new Application Message to publish,
		q.mut.Lock()
		p.PacketId = q.getUnusedPacketId()
		q.sent[p.PacketId] = p
		q.mut.Unlock()

		// MUST send a PUBLISH Packet containing this Packet Identifier
		// with QoS=1, DUP=0.
		p.Duplicate = false
		q.wrQueue.Add(&writequeue.Item{
			Packet: p,
		})

	case packet.QoS2:
		q.log.Error("Sending PUBLISH with QoS > 1 is not implemented")
		return errors.New("QoS above 1 is not implemented")
	}
	return nil
}
