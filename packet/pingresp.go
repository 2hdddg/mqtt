package packet

type PingResp struct {
}

func (r *Reader) readPingResp(fixflags uint8) (*PingResp, error) {
	return &PingResp{}, nil
}

func (p *PingResp) toPacket() []byte {
	return []byte{
		uint8(PINGRESP << 4),
		0,
	}
}

