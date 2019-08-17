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
	// that is identical to an existing Subscription’s Topic Filter then
	// it MUST completely replace that existing Subscription with a new
	// Subscription. The Topic Filter in the new Subscription will be
	// identical to that in the previous Subscription, although its
	// maximum QoS value could be different. Any existing retained
	// messages matching the Topic Filter MUST be re-sent, but the flow
	// of publications MUST NOT be interrupted [MQTT-3.8.4-3].
	// TODO: [MQTT-3.8.4-3], resend retained when replacing existing.
	// TODO: Replace
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

func (s *subscriptions) match(t *topic.Name) *subscription {
	// When Clients make subscriptions with Topic Filters that include
	// wildcards, it is possible for a Client’s subscriptions to
	// overlap so that a published message might match multiple filters.
	// In this case the Server MUST deliver the message to the Client
	// respecting the maximum QoS of all the matching subscriptions
	// [MQTT-3.3.4-2]. In addition, the Server MAY deliver further
	// copies of the message, one for each additional matching
	// subscription and respecting the subscription’s QoS in each case.

	// Version 5
	// If the Client specified a Subscription Identifier for any of the
	// overlapping subscriptions the Server MUST send those Subscription
	// Identifiers in the message which is published as the result of the
	// subscriptions [MQTT-3.3.4-3]. If the Server sends a single copy of
	// the message it MUST include in the PUBLISH packet the Subscription
	// Identifiers for all matching subscriptions which have a
	// Subscription Identifiers, their order is not significant
	// [MQTT-3.3.4-4]. If the Server sends multiple PUBLISH packets
	// it MUST send, in each of them, the Subscription Identifier of the
	// matching subscription if it has a Subscription Identifier
	// [MQTT-3.3.4-5].
	var matched *subscription
	for i, _ := range s.subs {
		sub := &s.subs[i]
		if sub.filter.Match(t) {
			if matched == nil || sub.qoS > matched.qoS {
				matched = sub
			}
		}
	}
	return matched
}
