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
		}},
		{&PingReq{}},
		{&PingResp{}},
		{&Disconnect{Reserved: 7}},
	}

	for _, c := range testcases {
		buf := bytes.NewBuffer([]byte{})
		wr := Writer{buf}
		err := wr.WritePacket(c.x1)
		if err != nil {
			t.Errorf("Failed to write packet")
		}
		rd := Reader{bufio.NewReader(buf)}
		x2, err := rd.ReadPacket(4)
		if err != nil {
			t.Errorf("Failed to read packet")
		}

		if !reflect.DeepEqual(c.x1, x2) {
			t.Errorf("Structs differ!")
		}
	}
}

