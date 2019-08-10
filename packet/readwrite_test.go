package packet

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestWriteReadPacket(t *testing.T) {
	testcases := []struct {
		x1 interface{}
	}{
		{&Connect{
			ProtocolName:     "MQTT",
			ProtocolVersion:  4,
			KeepAliveSecs:    30,
			ClientIdentifier: "1",
		}},
		{&Publish{
			Duplicate: true,
			QoS:       QoS1,
			Retain:    true,
			Topic:     "the topic",
			PacketId:  0xfffe,
			Payload:   []byte{0xfe, 0xff, 01, 02},
		}},
		{&PingReq{}},
		{&PingResp{}},
		{&Disconnect{Reserved: 7}},
		{&Subscribe{
			PacketId: 0x0107,
			Subscriptions: []Subscription{
				{Topic: "x/y/z", QoS: 2},
				{Topic: "a/#", QoS: 2},
			},
		}},
		{&SubscribeAck{
			PacketId: 0x1234,
			ReturnCodes: []QoS{
				QoS1, QoS0, QoS2, QoSFailure,
			},
		}},
	}

	for _, c := range testcases {
		buf := bytes.NewBuffer([]byte{})
		wr := Writer{buf}
		err := wr.WritePacket(c.x1)
		if err != nil {
			t.Errorf("Failed to write packet: %s", err)
		}
		rd := Reader{bufio.NewReader(buf)}
		x2, err := rd.ReadPacket(4)
		if err != nil {
			t.Errorf("Failed to read packet: %s", err)
		}

		if !reflect.DeepEqual(c.x1, x2) {
			t.Errorf("Structs differ!")
		}
	}
}

