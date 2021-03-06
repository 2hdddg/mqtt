package packet

type QoS uint8

// QoS 0: At most once delivery
// QoS 1: At least once delivery
// QoS 2: Exactly once delivery
const (
	QoS0       QoS = 0x00
	QoS1       QoS = 0x01
	QoS2       QoS = 0x02
	QoSFailure QoS = 0x80
)

var QoSHighest = QoS2

type Connect struct {
	ProtocolName     string
	ProtocolVersion  uint8
	WillRetain       bool
	WillQoS          QoS
	CleanStart       bool
	KeepAliveSecs    uint16
	ClientIdentifier string
	WillTopic        *string
	WillMessage      []byte
	UserName         *string
	Password         []byte
}

func (r *Reader) readConnect(fixflags uint8) (*Connect, error) {
	var err error
	c := &Connect{}

	c.ProtocolName, err = r.str()
	if err != nil {
		return nil, err
	}
	c.ProtocolVersion, err = r.ReadByte()
	if err != nil {
		return nil, err
	}

	conflags, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	conflags = conflags >> 1
	c.CleanStart = (conflags & 0x01) > 0
	conflags = conflags >> 1
	willFlag := (conflags & 0x01) > 0
	conflags = conflags >> 1
	if willFlag {
		qoS := conflags & 0x03
		if qoS > 3 {
			return nil, newProtoErr("Illegal QoS value")
		}
		c.WillQoS = QoS(qoS)
		conflags = conflags >> 2
		c.WillRetain = (conflags & 0x01) > 0
		conflags = conflags >> 1
	} else {
		c.WillQoS = QoS0
		c.WillRetain = false
		conflags = conflags >> 3
	}
	passwordFlag := (conflags & 0x01) > 0
	conflags = conflags >> 1
	userNameFlag := (conflags & 0x01) > 0

	c.KeepAliveSecs, err = r.int2()
	if err != nil {
		return nil, err
	}

	if c.ProtocolVersion > 4 {
		// TODO:
		// Read connect properties
	}

	// Read Client identifier
	c.ClientIdentifier, err = r.str()
	if err != nil {
		return nil, err
	}

	if willFlag {
		if c.ProtocolVersion > 4 {
			// TODO:
			// Read will properties
		}

		// Will topic
		willTopic, err := r.str()
		if err != nil {
			return nil, err
		}
		c.WillTopic = &willTopic

		// Will message
		c.WillMessage, err = r.bin()
		if err != nil {
			return nil, err
		}
	}

	if userNameFlag {
		// TODO:
	}

	if passwordFlag {
		// TODO:
	}

	return c, nil
}

func (c *Connect) toPacket() []byte {
	// Variable header
	v := strToBytes(c.ProtocolName)
	v = append(v, c.ProtocolVersion)
	flags := flagsToBitsU8([]bool{
		false,
		c.CleanStart,
		c.WillMessage != nil})
	flags = flags | (uint8(c.WillQoS) << 3)
	flags = flags | (flagsToBitsU8([]bool{
		c.WillRetain,
		c.Password != nil,
		c.UserName != nil}) << 5)
	v = append(v, flags)
	v = append(v, toInt2(c.KeepAliveSecs)...)
	// Payload
	v = append(v, strToBytes(c.ClientIdentifier)...)
	if c.WillMessage != nil {
		// TODO:
	}
	if c.UserName != nil {
		// TODO:
	}
	if c.Password != nil {
		// TODO:
	}

	rem := toVarInt(uint32(len(v)))
	h := make([]uint8, 1, len(v)+len(rem)+1)
	h[0] = uint8(CONNECT << 4)
	h = append(h, rem...)
	h = append(h, v...)

	return h
}

func (c *Connect) name() string {
	return "CONNECT"
}
