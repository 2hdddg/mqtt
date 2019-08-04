package controlpacket

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

