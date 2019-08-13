package packet

import "io"

type SubscribeAck struct {
	PacketId    uint16
	ReturnCodes []QoS
}

func (r *Reader) readSubscribeAck(fixflags uint8) (*SubscribeAck, error) {
	var err error
	s := &SubscribeAck{}

	// From variable header
	s.PacketId, err = r.int2()
	if err != nil {
		return nil, err
	}

	// Returns codes for N number of subscriptions
	for {
		qoS, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if qoS != byte(QoSFailure) && qoS > byte(QoSHighest) {
			return nil, newProtoErr("Illegal QoS")
		}
		s.ReturnCodes = append(s.ReturnCodes, QoS(qoS))
	}

	return s, nil
}

func (s *SubscribeAck) toPacket() []byte {
	v := toInt2(s.PacketId)

	// Payload, N number of subscriptions return codes
	for _, qoS := range s.ReturnCodes {
		v = append(v, byte(qoS))
	}

	rem := toVarInt(uint32(len(v)))
	h := make([]uint8, 1, len(v)+len(rem)+1)
	h[0] = uint8(SUBACK<<4) | 2
	h = append(h, rem...)
	h = append(h, v...)

	return h
}
