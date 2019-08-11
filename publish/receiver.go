package publish

import (
	"errors"

	"github.com/2hdddg/mqtt/packet"
)

type Accept func(p *packet.Publish)
type AwaitAck func(p packet.Packet, packetId uint16)

type Receiver struct {
	packets  map[uint16]*packet.Publish
	accept   Accept
	awaitAck AwaitAck
}

// Funcs execute on same thread as state transition functions are
// called on.
func NewReceiver(accept Accept, awaitAck AwaitAck) *Receiver {
	return &Receiver{
		accept:   accept,
		awaitAck: awaitAck,
		packets:  make(map[uint16]*packet.Publish),
	}
}

func (r *Receiver) Received(p *packet.Publish) error {
	switch p.QoS {
	case packet.QoS0:
		r.accept(p)
	case packet.QoS1:
		r.packets[p.PacketId] = p
		ack := &packet.PublishAck{PacketId: p.PacketId}
		r.awaitAck(ack, p.PacketId)
	case packet.QoS2:
		return errors.New("QoS above 1 is not implemented")
	}
	return nil
}

func (r *Receiver) Ack(packetId uint16) {
	p, exists := r.packets[packetId]
	if exists && p != nil {
		r.accept(p)
		delete(r.packets, packetId)
	}
}
