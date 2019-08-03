package controlpacket

import (
	"bytes"
	"testing"
)

func TestFlagsToBitsU8(t *testing.T) {
	testcases := []struct {
		flags []bool
		bits  uint8
	}{
		{[]bool{}, 0},
		{[]bool{true}, 0x01},
		{[]bool{true, false}, 0x02},
		{[]bool{true, false, true, false}, 10},
	}

	for _, c := range testcases {
		bits := flagsToBitsU8(c.flags)
		if bits != c.bits {
			t.Errorf("Expected %d but was %d", c.bits, bits)
		}
	}
}

func TestToVarInt(t *testing.T) {
	testcases := []struct {
		val    uint32
		varint []byte
	}{
		{0, []byte{0x00}},
		{0x7f, []byte{0x7f}},
		{0x80, []byte{0x80, 0x01}},
		{16383, []byte{0xff, 0x7f}},
		{16384, []byte{0x80, 0x80, 0x01}},
		{2097151, []byte{0xff, 0xff, 0x7f}},
	}

	for _, c := range testcases {
		varint := toVarInt(c.val)
		if !bytes.Equal(varint, c.varint) {
			t.Errorf("Expected %d but was %d", c.varint, varint)
		}
	}
}
