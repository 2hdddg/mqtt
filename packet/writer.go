package packet

import (
	"errors"
	"io"
)

type Writer struct {
	io.Writer
}

type packetize interface {
	toPacket() []byte
}

func (w *Writer) WritePacket(packet interface{}) error {
	p, ok := packet.(packetize)
	if !ok {
		return errors.New("Wrong type")
	}
	b := p.toPacket()
	_, err := w.Write(b)
	return err
}
