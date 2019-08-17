package server

import (
	"reflect"
	"testing"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

func TestSubscribe(t *testing.T) {
	testcases := []struct {
		sub packet.Subscription
		qoS packet.QoS
	}{
		// Good topic and QoS
		{
			packet.Subscription{
				QoS:   packet.QoS0,
				Topic: "x/y",
			},
			packet.QoS0,
		},
		// Bad topic
		{
			packet.Subscription{
				QoS:   packet.QoS0,
				Topic: "",
			},
			packet.QoSFailure,
		},
	}

	for _, c := range testcases {
		s := newSubscriptions()
		qoS := s.subscribe(&c.sub)
		if qoS != c.qoS {
			t.Errorf("Expected QoS %v but got %v", c.qoS, qoS)
		}
	}
}

func TestMatch(t *testing.T) {
	testcases := []struct {
		subs    []packet.Subscription
		topic   *topic.Name
		matched *subscription
	}{
		// One match
		{
			[]packet.Subscription{
				packet.Subscription{
					QoS:   packet.QoS0,
					Topic: "x/y",
				},
			},
			topic.NewName("x/y"),
			&subscription{
				qoS:    packet.QoS0,
				filter: *topic.NewFilter("x/y"),
			},
		},
		// No match
		{
			[]packet.Subscription{
				packet.Subscription{
					QoS:   packet.QoS0,
					Topic: "x/y",
				},
			},
			topic.NewName("x/z"),
			nil,
		},
		// Two matches, pick highest QoS
		{
			[]packet.Subscription{
				packet.Subscription{
					QoS:   packet.QoS0,
					Topic: "x/y",
				},
				packet.Subscription{
					QoS:   packet.QoS2,
					Topic: "x/#",
				},
			},
			topic.NewName("x/y"),
			&subscription{
				qoS:    packet.QoS2,
				filter: *topic.NewFilter("x/#"),
			},
		},
	}
	for _, c := range testcases {
		s := newSubscriptions()
		for _, sub := range c.subs {
			s.subscribe(&sub)
		}
		matched := s.match(c.topic)
		if !reflect.DeepEqual(matched, c.matched) {
			t.Errorf("Matching differs")
		}
	}
}
