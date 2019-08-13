package conn

import (
	"time"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"

)

type C interface {
	ReadPacket(version uint8, log logger.L) (packet.Packet, error)
	WritePacket(packet packet.Packet, log logger.L) error
	Close() error
	SetReadDeadline(t time.Time) error
}

