package server

import (
	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

type subscription struct {
	filter topic.Filter
	qoS    packet.QoS
}

type subscriptions struct {
	subs []subscription
}

func newSubscriptions() *subscriptions {
	return &subscriptions{
		subs: make([]subscription, 0),
	}
}

func (s *subscriptions) subscribe(sub *packet.Subscription) packet.QoS {
	// If a Server receives a SUBSCRIBE Packet containing a Topic Filter
	// that is identical to an existing Subscription’s Topic Filter then it
	// MUST completely replace that existing Subscription with a new
	// Subscription. The Topic Filter in the new Subscription will be
	// identical to that in the previous Subscription, although its maximum
	// QoS value could be different. Any existing retained messages
	// matching the Topic Filter MUST be re-sent, but the flow of
	// publications MUST NOT be interrupted [MQTT-3.8.4-3].
	// TODO: [MQTT-3.8.4-3], resend retained when replacing existing.
	filter := topic.NewFilter(sub.Topic)
	if filter != nil {
		s.subs = append(s.subs, subscription{
			filter: *filter,
			qoS:    sub.QoS,
		})
		return sub.QoS
	} else {
		return packet.QoSFailure
	}
}

func (s *subscriptions) match(t *topic.Name) (bool, packet.QoS) {
	qoS := -1
	for _, sub := range s.subs {
		if sub.filter.Match(t) {
			if int(sub.qoS) > qoS {
				qoS = int(sub.qoS)
			}
		}
	}
	if qoS == -1 {
		return false, packet.QoSFailure
	}
	return true, packet.QoS(qoS)
}
