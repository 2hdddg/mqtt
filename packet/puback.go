package packet

type PublishAck struct {
	PacketId uint16
}

func (r *Reader) readPublishAck(fixflags uint8) (*PublishAck, error) {
	var err error
	p := &PublishAck{}

	// From variable header
	p.PacketId, err = r.int2()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PublishAck) toPacket() []byte {
	h := make([]byte, 4)
	h[0] = uint8(PUBACK << 4)
	h[1] = 0x02
	h[2] = uint8(p.PacketId >> 8)
	h[3] = uint8(p.PacketId)
	return h
}
