package packet

import (
	"io"
)

type Writer struct {
	io.Writer
}

type packetize interface {
	toPacket() []byte
}

func (w *Writer) write(p packetize) error {
	packet := p.toPacket()
	_, err := w.Write(packet)
	return err
}

func (w *Writer) WriteConnect(x *Connect) error {
	return w.write(x)
}

func (w *Writer) WriteAckConnection(x *AckConnection) error {
	return w.write(x)
}

func (w *Writer) WritePingResp(x *PingResp) error {
	return w.write(x)
}

func (w *Writer) WritePublish(x *Publish) error {
	return w.write(x)
}
