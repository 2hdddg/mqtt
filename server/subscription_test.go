package server

import (
	"testing"

	"github.com/2hdddg/mqtt/packet"
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
			t.Errorf("Expected %v but got %v", c.qoS, qoS)
		}
	}
}
