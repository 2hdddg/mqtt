package packet

import (
	"io"

	"github.com/2hdddg/mqtt/logger"
)

type Writer struct {
	io.Writer
}

func (w *Writer) WritePacket(packet Packet, log logger.L) error {
	b := packet.toPacket()
	n, err := w.Write(packet.toPacket())
	if err != nil {
		return err
	}
	if n != len(b) {
		return newProtoErr("Wrote too few bytes")
	}
	log.Info("Sent " + packet.name())
	return nil
}
