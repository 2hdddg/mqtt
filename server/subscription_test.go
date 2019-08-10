package server

import (
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
		qoS     packet.QoS
		matched bool
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
			packet.QoS0,
			true,
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
			packet.QoSFailure,
			false,
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
			packet.QoS2, // Should get highest!
			true,
		},
	}
	for _, c := range testcases {
		s := newSubscriptions()
		for _, sub := range c.subs {
			s.subscribe(&sub)
		}
		matched, qoS := s.match(c.topic)
		if matched != c.matched {
			t.Errorf("Expected matched %v but was %v",
				c.matched, matched)
		}
		if qoS != c.qoS {
			t.Errorf("Expected QoS %v but got %v", c.qoS, qoS)
		}
	}
}
