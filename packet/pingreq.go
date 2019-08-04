package packet

type PingReq struct {
}

func (r *Reader) readPingReq(fixflags uint8) (*PingReq, error) {
	return &PingReq{}, nil
}
