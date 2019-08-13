package packet

type PingReq struct {
}

func (r *Reader) readPingReq(fixflags uint8) (*PingReq, error) {
	return &PingReq{}, nil
}

func (p *PingReq) toPacket() []byte {
	return []byte{
		uint8(PINGREQ << 4),
		0,
	}
}

func (c *PingReq) name() string {
	return "PINGREQ"
}
