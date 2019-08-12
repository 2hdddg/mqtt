package publish

import (
	"errors"
	"sync"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/writequeue"
)

type Accept func(p *packet.Publish)

type Logger interface {
	Info(s string)
	Error(s string)
	Debug(s string)
}

type Receiver struct {
	packets map[uint16]*packet.Publish
	accept  Accept
	wrQueue *writequeue.Queue
	mut     *sync.Mutex
	log     Logger
}

func NewReceiver(
	accept Accept, wrQueue *writequeue.Queue,
	log Logger) *Receiver {

	return &Receiver{
		accept:  accept,
		wrQueue: wrQueue,
		packets: make(map[uint16]*packet.Publish),
		mut:     &sync.Mutex{},
		log:     log,
	}
}

func (r *Receiver) written(packetId uint16) {
	r.mut.Lock()
	p, exists := r.packets[packetId]
	if exists && p != nil {
		go r.accept(p)
		delete(r.packets, packetId)
	}
	r.mut.Unlock()
}

func (r *Receiver) Received(p *packet.Publish) error {
	switch p.QoS {
	case packet.QoS0:
		r.accept(p)
	case packet.QoS1:
		r.mut.Lock()
		r.packets[p.PacketId] = p
		r.mut.Unlock()
		ack := &packet.PublishAck{PacketId: p.PacketId}
		r.wrQueue.Add(&writequeue.Item{
			Packet: ack,
			Written: func() {
				r.written(p.PacketId)
				r.log.Info("Sent PUBACK")
			},
		})
	case packet.QoS2:
		return errors.New("QoS above 1 is not implemented")
	}
	return nil
}
