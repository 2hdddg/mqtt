package packet

import (
	"bufio"
	"bytes"
	"fmt"

	"io"
)

type Type uint8

const (
	CONNECT     Type = 1
	CONNACK     Type = 2
	PUBLISH     Type = 3
	PUBACK      Type = 4
	PUBREC      Type = 5
	PUBREL      Type = 6
	PUBCOMP     Type = 7
	SUBSCRIBE   Type = 8
	SUBACK      Type = 9
	UNSUBSCRIBE Type = 10
	UNSUBACK    Type = 11
	PINGREQ     Type = 12
	PINGRESP    Type = 13
	DISCONNECT  Type = 14
	AUTH        Type = 15
)

func (t Type) String() string {
	switch t {
	case CONNECT:
		return "CONNECT"
	case CONNACK:
		return "CONNACK"
	case PUBLISH:
		return "PUBLISH"
	case PUBACK:
		return "PUBACK"
	case PUBREC:
		return "PUBREC"
	case PUBREL:
		return "PUBREL"
	case PUBCOMP:
		return "PUBCOMP"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBACK:
		return "SUBACK"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case UNSUBACK:
		return "UNSUBACK"
	case PINGREQ:
		return "PINGREQ"
	case PINGRESP:
		return "PINGRESP"
	case DISCONNECT:
		return "DISCONNECT"
	case AUTH:
		return "AUTH"
	}
	return "<unknown>"
}

type Reader struct {
	*bufio.Reader
}

func (r *Reader) ReadPacket(version uint8) (interface{}, error) {
	// Read fixed header
	ctrlAndFlags, err := r.ReadByte()
	if err != nil {
		return nil,
			&Error{c: "fix header", m: "control packet type", err: err}
	}

	// Extract packet type and flags
	t := Type(ctrlAndFlags >> 4)
	f := ctrlAndFlags & 0x0f

	fmt.Printf("Received %s\n", t)

	// Read remaining length
	rem, err := r.varInt()
	if err != nil {
		return nil, err
	}

	// Read the rest
	buf := make([]byte, rem)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, &Error{c: "control packet read",
			m: "failed to fill buffer", err: err}
	}
	r = &Reader{bufio.NewReader(bytes.NewBuffer(buf))}

	var p interface{}
	switch t {
	case CONNECT:
		p, err = r.readConnect(f)
	case DISCONNECT:
		p, err = r.readDisconnect(f)
	case PUBLISH:
		p, err = r.readPublish(f)
	case PINGREQ:
		p, err = r.readPingReq(f)
	default:
		err = &Error{c: "control packet read",
			m: "Unhandled packet", err: err}
	}
	return p, err
}