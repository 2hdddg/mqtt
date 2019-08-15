package client

import (
	"fmt"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

func Connect(conn conn.C, log logger.L) bool {
	c := &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		CleanStart:       true,
		KeepAliveSecs:    3,
		ClientIdentifier: "hoho",
	}
	err := conn.WritePacket(c, log)
	if err != nil {
		return false
	}

	ack, err := conn.ReadPacket(c.ProtocolVersion, log)
	if err != nil {
		return false
	}
	fmt.Println(ack)
	return true
}
