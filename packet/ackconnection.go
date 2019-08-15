package packet

type ConnRetCode uint8

const (
	ConnAccepted                     ConnRetCode = 0
	ConnRefusedVersion               ConnRetCode = 1
	ConnRefusedIdentifier            ConnRetCode = 2
	ConnRefusedServerUnavailable     ConnRetCode = 3
	ConnRefusedBadUsernameOrPassword ConnRetCode = 4
	ConnRefusedNotAuthorized         ConnRetCode = 5
)

type AckConnection struct {
	SessionPresent bool
	RetCode        ConnRetCode
}

func RefuseConnection(retCode ConnRetCode) *AckConnection {
	// If a server sends a CONNACK packet containing a non-zero
	// return code it MUST set Session Present to 0
	if retCode == ConnAccepted {
		return nil
	}
	return &AckConnection{
		SessionPresent: false,
		RetCode:        retCode,
	}
}

func (a *AckConnection) toPacket() []byte {
	ctrlAndFlags := uint8(CONNACK << 4)
	ackFlags := flagsToBitsU8([]bool{a.SessionPresent})

	return []byte{
		ctrlAndFlags,
		2,
		ackFlags,
		uint8(a.RetCode),
	}
}

func (r *Reader) readConnectAck(fixflags uint8) (*AckConnection, error) {
	var err error
	c := &AckConnection{}
	ackFlags, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if ackFlags > 1 {
		return nil, newProtoErr("Illegal ack flags")
	}
	c.SessionPresent = (ackFlags & 0x01) == 1

	retCode, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	c.RetCode = ConnRetCode(retCode)
	return c, nil
}

func (a *AckConnection) name() string {
	return "CONNACK"
}

