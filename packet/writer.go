package packet

import (
	"errors"
	"io"

	"github.com/2hdddg/mqtt/logger"
)

type Writer struct {
	io.Writer
}

type packetize interface {
	toPacket() []byte
}

func (w *Writer) WritePacket(packet Packet, log logger.L) error {
	p, ok := packet.(packetize)
	if !ok {
		log.Error("Trying to write unknown packet type")
		return errors.New("Wrong type")
	}
	b := p.toPacket()
	_, err := w.Write(b)
	return err
}
