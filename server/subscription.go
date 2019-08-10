package server

import (
	"fmt"

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
		fmt.Println("Evaluating subscription", sub)
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
