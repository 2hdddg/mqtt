package packet

import (
	"bufio"
	"bytes"
	"testing"
)

func TestInt2(t *testing.T) {
	testcases := []struct {
		buf []byte
		val uint16
		err bool
	}{
		{[]byte{0, 0}, 0, false},
		{[]byte{0, 4}, 4, false},
	}
	for _, c := range testcases {
		r := &Reader{bufio.NewReader(bytes.NewBuffer(c.buf))}
		val, err := r.int2()
		if c.err && err == nil {
			t.Errorf("Expected error but didn't get it")
			continue
		}
		if !c.err && err != nil {
			t.Errorf("Didn't expect error but got %s", err)
			continue
		}
		if val != c.val {
			t.Errorf("Expected val %d but was %d", c.val, val)
		}
	}
}

func TestVarInt(t *testing.T) {
	testcases := []struct {
		buf []byte
		val uint32
		err bool
	}{
		{[]byte{0}, 0, false},
		{[]byte{0x7f}, 127, false},
		{[]byte{0x80, 0x01}, 128, false},
		{[]byte{0xff, 0x7f}, 16383, false},
		{[]byte{0x80, 0x80, 0x01}, 16384, false},
		{[]byte{0xff, 0xff, 0x7f}, 2097151, false},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152, false},
		{[]byte{0xff, 0xff, 0xff, 0x7f}, 268435455, false},
		{[]byte{}, 0, true},
		{[]byte{0xff}, 0, true},
		{[]byte{0xff, 0xff}, 0, true},
		{[]byte{0xff, 0xff, 0x80, 0xff, 0x01}, 0, true},
	}

	for _, c := range testcases {
		r := &Reader{bufio.NewReader(bytes.NewBuffer(c.buf))}
		val, err := r.varInt()
		if c.err && err == nil {
			t.Errorf("Expected error but didn't get it")
			continue
		}
		if !c.err && err != nil {
			t.Errorf("Didn't expect error but got %s", err)
			continue
		}
		if val != c.val {
			t.Errorf("Expected val %d but was %d", c.val, val)
		}
	}
}

func TestString(t *testing.T) {
	testcases := []struct {
		buf []byte
		val string
		err bool
	}{
		{[]byte{0, 4, 'M', 'Q', 'T', 'T'}, "MQTT", false},
	}
	for _, c := range testcases {
		r := &Reader{bufio.NewReader(bytes.NewBuffer(c.buf))}
		val, err := r.str()
		if c.err && err == nil {
			t.Errorf("Expected error but didn't get it")
			continue
		}
		if !c.err && err != nil {
			t.Errorf("Didn't expect error but got %s", err)
			continue
		}
		if val != c.val {
			t.Errorf("Expected val %s but was %s", c.val, val)
		}
	}
}

func TestBinary(t *testing.T) {
	testcases := []struct {
		buf []byte
		val []byte
		err bool
	}{
		{[]byte{0, 4, 'M', 'Q', 'T', 'T'},
			[]byte{'M', 'Q', 'T', 'T'}, false},
	}
	for _, c := range testcases {
		r := &Reader{bufio.NewReader(bytes.NewBuffer(c.buf))}
		val, err := r.bin()
		if c.err && err == nil {
			t.Errorf("Expected error but didn't get it")
			continue
		}
		if !c.err && err != nil {
			t.Errorf("Didn't expect error but got %s", err)
			continue
		}
		if !bytes.Equal(val, c.val) {
			t.Errorf("Expected val %s but was %s", c.val, val)
		}
	}
}
