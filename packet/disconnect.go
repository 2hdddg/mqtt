package packet

type Disconnect struct {
	Reserved uint8
}

func (d *Disconnect) toPacket() []byte {
	return []byte{
		uint8(DISCONNECT<<4) | d.Reserved&0x0f,
		0,
	}
}

func (r *Reader) readDisconnect(fixflags uint8) (*Disconnect, error) {
	// TODO: The Server MUST validate that reserved bits are set to zero
	// and disconnect the Client if they are not zero
	return &Disconnect{Reserved: fixflags}, nil
}

func (c *Disconnect) name() string {
	return "DISCONNECT"
}
