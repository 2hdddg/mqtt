package packet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/2hdddg/mqtt/logger"
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

type Packet interface {
	toPacket() []byte
	name() string
}

func (r *Reader) ReadPacket(version uint8, log logger.L) (Packet, error) {
	// Read fixed header
	log.Debug("Waiting for data to read")
	ctrlAndFlags, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	// Extract packet type and flags
	t := Type(ctrlAndFlags >> 4)
	f := ctrlAndFlags & 0x0f

	// Read remaining length
	rem, err := r.varInt()
	if err != nil {
		return nil, err
	}

	// Read the rest
	buf := make([]byte, rem)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	r = &Reader{bufio.NewReader(bytes.NewBuffer(buf))}

	var p Packet
	switch t {
	case CONNECT:
		p, err = r.readConnect(f)
	case CONNACK:
		p, err = r.readConnectAck(f)
	case DISCONNECT:
		p, err = r.readDisconnect(f)
	case PUBLISH:
		p, err = r.readPublish(f)
	case PUBACK:
		p, err = r.readPublishAck(f)
	case PINGREQ:
		p, err = r.readPingReq(f)
	case PINGRESP:
		p, err = r.readPingResp(f)
	case SUBSCRIBE:
		p, err = r.readSubscribe(f)
	case SUBACK:
		p, err = r.readSubscribeAck(f)
	}
	if err != nil {
		return nil, err
	}
	if p != nil {
		log.Info(fmt.Sprintf("Received %s", p.name()))
	} else {
		m := fmt.Sprintf("Unhandled packet type %d", t)
		log.Error(m)
		return nil, newProtoErr(m)
	}
	return p, nil
}
