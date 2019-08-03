package controlpacket

import (
	"bufio"
	"bytes"
	"testing"
)

func TestReadPacketConnect(t *testing.T) {
	c1 := Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		KeepAliveSecs:    30,
		ClientIdentifier: "1",
	}
	b := c1.toPacket()
	r := Reader{bufio.NewReader(bytes.NewBuffer(b))}
	cx, err := r.ReadPacket(4)
	if err != nil {
		t.Fatalf("Failed to read packet: %s", err)
	}
	if cx == nil {
		t.Errorf("Didn't get a packet")
	}
	c2 := cx.(*Connect)
	if c1.ProtocolName != c2.ProtocolName ||
		c1.ProtocolVersion != c2.ProtocolVersion ||
		c1.KeepAliveSecs != c2.KeepAliveSecs ||
		c1.ClientIdentifier != c2.ClientIdentifier {
		t.Errorf("Packet differs!")
	}
}

func TestReadPacketPublish(t *testing.T) {
	c1 := Publish{
		Duplicate: true,
		QoS:       QoS1,
		Retain:    true,
		Topic:     "the topic",
		PacketId:  0xfffe,
	}
	b := c1.toPacket()
	r := Reader{bufio.NewReader(bytes.NewBuffer(b))}
	cx, err := r.ReadPacket(4)
	if err != nil {
		t.Fatalf("Failed to read packet: %s", err)
	}
	if cx == nil {
		t.Errorf("Didn't get a packet")
	}
	c2 := cx.(*Publish)
	if c1.Duplicate != c2.Duplicate ||
		c1.QoS != c2.QoS ||
		c1.Retain != c2.Retain ||
		c1.Topic != c2.Topic ||
		c1.PacketId != c2.PacketId {
		t.Errorf("Packet differs: %v vs %v", c1, c2)
	}
}
